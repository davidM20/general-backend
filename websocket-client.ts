// WebSocket Client Library para conectar con el servidor customws de Go
// Compatible con el protocolo genérico definido en el backend

export type MessageType =
    // Tipos de mensajes Cliente -> Servidor (genéricos)
    | "data_request"
    | "presence_update"
    | "client_ack"
    | "generic_request"
    // Tipos de mensajes Servidor -> Cliente (genéricos)
    | "data_event"
    | "presence_event"
    | "server_ack"
    | "generic_response"
    | "error_notification";

export interface ClientToServerMessage {
    pid?: string;
    type: MessageType;
    targetUserId?: number;
    payload: any;
}

export interface ServerToClientMessage {
    pid?: string;
    type: MessageType;
    fromUserId?: number;
    payload: any;
    error?: ErrorPayload;
}

export interface ErrorPayload {
    originalPid?: string;
    code?: number;
    message: string;
}

export interface AckPayload {
    acknowledgedPid: string;
    status?: string;
    error?: string;
}

export interface PresenceUpdatePayload {
    status: string;
    targetUserId?: number;
}

// Payload genérico para data_request usando map[string]interface{} del backend
export interface DataRequestPayload {
    action?: string;
    resource?: string;
    data?: Record<string, any>;
    [key: string]: any; // Permite cualquier campo adicional
}

// Ejemplos de payloads específicos que pueden usar los desarrolladores
export interface ChatMessagePayload extends DataRequestPayload {
    action: "send_message" | "get_history" | "mark_read";
    resource: "chat";
    data: {
        text?: string;
        messageId?: string;
        chatId?: string;
        timestamp?: string;
    };
}

export interface FileUploadPayload extends DataRequestPayload {
    action: "upload_chunk" | "complete_upload" | "cancel_upload";
    resource: "file";
    data: {
        fileName: string;
        fileSize: number;
        mimeType: string;
        chunkNum: number;
        totalChunks: number;
        fileId: string;
        data: string; // Base64 encoded chunk
    };
}

export interface NotificationPayload extends DataRequestPayload {
    action: "send" | "mark_read" | "get_pending";
    resource: "notification";
    data: {
        type?: string;
        title?: string;
        message?: string;
        urgent?: boolean;
        targetUsers?: number[];
        expireAt?: string;
    };
}

export interface WSClientConfig {
    url: string;
    authHeaders?: Record<string, string>;
    authParams?: URLSearchParams;
    reconnectInterval?: number;
    maxReconnectAttempts?: number;
    ackTimeout?: number;
    requestTimeout?: number;
    pingInterval?: number;
    enableDebugLogs?: boolean;
}

export interface PendingAck {
    resolve: (ack: ServerToClientMessage) => void;
    reject: (error: Error) => void;
    timestamp: number;
    messageId: string;
}

export interface PendingResponse {
    resolve: (response: ServerToClientMessage) => void;
    reject: (error: Error) => void;
    timestamp: number;
}

export type ConnectionState = 'connecting' | 'connected' | 'disconnected' | 'reconnecting' | 'failed';

export interface WSClientCallbacks<TUserData = any> {
    onConnect?: (userData?: TUserData) => void;
    onDisconnect?: (error?: Error) => void;
    onMessage?: (message: ServerToClientMessage) => void;
    onDataEvent?: (message: ServerToClientMessage) => void;
    onPresenceEvent?: (message: ServerToClientMessage) => void;
    onErrorNotification?: (message: ServerToClientMessage) => void;
    onConnectionStateChange?: (state: ConnectionState) => void;
}

export class WebSocketClient<TUserData = any> {
    private ws: WebSocket | null = null;
    private config: Required<WSClientConfig>;
    private callbacks: WSClientCallbacks<TUserData>;
    private reconnectAttempts = 0;
    private reconnectTimer: number | null = null;
    private pingTimer: number | null = null;
    private cleanupTimer: number | null = null;
    private connectionState: ConnectionState = 'disconnected';

    // Mapas para manejar ACKs y respuestas pendientes
    private pendingAcks = new Map<string, PendingAck>();
    private pendingResponses = new Map<string, PendingResponse>();

    // Contador para generar PIDs únicos
    private pidCounter = 0;

    constructor(config: WSClientConfig, callbacks: WSClientCallbacks<TUserData> = {}) {
        this.config = {
            reconnectInterval: 5000,
            maxReconnectAttempts: 10,
            ackTimeout: 5000,
            requestTimeout: 10000,
            pingInterval: 30000,
            enableDebugLogs: false,
            authHeaders: {},
            authParams: new URLSearchParams(),
            ...config
        };
        this.callbacks = callbacks;

        // Iniciar rutina de limpieza
        this.startCleanupRoutine();
    }

    private generatePID(): string {
        return `client_${Date.now()}_${++this.pidCounter}`;
    }

    private log(message: string, data?: any): void {
        if (this.config.enableDebugLogs) {
            console.log(`[WSClient] ${message}`, data || '');
        }
    }

    private setConnectionState(state: ConnectionState): void {
        if (this.connectionState !== state) {
            this.connectionState = state;
            this.callbacks.onConnectionStateChange?.(state);
            this.log(`Estado de conexión cambiado a: ${state}`);
        }
    }

    public getConnectionState(): ConnectionState {
        return this.connectionState;
    }

    public connect(): Promise<void> {
        return new Promise((resolve, reject) => {
            if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                resolve();
                return;
            }

            this.setConnectionState('connecting');
            this.log('Iniciando conexión WebSocket...');

            try {
                // Construir URL con parámetros de autenticación si existen
                const url = new URL(this.config.url);
                if (this.config.authParams) {
                    this.config.authParams.forEach((value, key) => {
                        url.searchParams.set(key, value);
                    });
                }

                this.ws = new WebSocket(url.toString());

                // Agregar headers de autenticación si están disponibles
                // Nota: Los headers en WebSocket del navegador son limitados
                // La autenticación generalmente se hace vía token en URL o en el primer mensaje

                this.ws.onopen = () => {
                    this.log('Conexión WebSocket establecida');
                    this.setConnectionState('connected');
                    this.reconnectAttempts = 0;
                    this.startPingRoutine();
                    this.callbacks.onConnect?.();
                    resolve();
                };

                this.ws.onmessage = (event) => {
                    this.handleMessage(event.data);
                };

                this.ws.onclose = (event) => {
                    this.log(`Conexión WebSocket cerrada. Código: ${event.code}, Razón: ${event.reason}`);
                    this.handleDisconnection(new Error(`WebSocket cerrado: ${event.code} - ${event.reason}`));
                };

                this.ws.onerror = (event) => {
                    this.log('Error en WebSocket', event);
                    const error = new Error('Error de WebSocket');
                    this.handleDisconnection(error);
                    if (this.connectionState === 'connecting') {
                        reject(error);
                    }
                };

            } catch (error) {
                this.log('Error al crear WebSocket', error);
                this.setConnectionState('failed');
                reject(error);
            }
        });
    }

    private handleMessage(data: string): void {
        try {
            const message: ServerToClientMessage = JSON.parse(data);
            this.log(`Mensaje recibido - Tipo: ${message.type}, PID: ${message.pid}, FromUserId: ${message.fromUserId}`);

            // Manejar ACKs del servidor
            if (message.type === 'server_ack') {
                this.handleServerAck(message);
                return;
            }

            // Manejar respuestas a solicitudes genéricas
            if (message.pid && this.pendingResponses.has(message.pid)) {
                this.handlePendingResponse(message);
                return;
            }

            // Manejar diferentes tipos de mensajes
            switch (message.type) {
                case 'data_event':
                    this.callbacks.onDataEvent?.(message);
                    break;
                case 'presence_event':
                    this.callbacks.onPresenceEvent?.(message);
                    break;
                case 'error_notification':
                    this.callbacks.onErrorNotification?.(message);
                    break;
                case 'generic_response':
                    // Los generic_response se manejan en pendingResponses arriba
                    this.log(`Respuesta genérica recibida: ${message.pid}`);
                    break;
                default:
                    this.log(`Tipo de mensaje no manejado: ${message.type}`);
            }

            // Callback general para todos los mensajes
            this.callbacks.onMessage?.(message);

        } catch (error) {
            this.log('Error al parsear mensaje', error);
        }
    }

    private handleServerAck(message: ServerToClientMessage): void {
        const ackPayload = message.payload as AckPayload;
        const pendingAck = this.pendingAcks.get(ackPayload.acknowledgedPid);

        if (pendingAck) {
            this.pendingAcks.delete(ackPayload.acknowledgedPid);

            if (ackPayload.error) {
                pendingAck.reject(new Error(ackPayload.error));
            } else {
                pendingAck.resolve(message);
            }

            this.log(`Server ACK procesado para PID: ${ackPayload.acknowledgedPid}`);
        } else {
            this.log(`Recibido Server ACK para PID desconocido: ${ackPayload.acknowledgedPid}`);
        }
    }

    private handlePendingResponse(message: ServerToClientMessage): void {
        const pendingResponse = this.pendingResponses.get(message.pid!);

        if (pendingResponse) {
            this.pendingResponses.delete(message.pid!);
            pendingResponse.resolve(message);
            this.log(`Respuesta procesada para PID: ${message.pid}`);
        }
    }

    private handleDisconnection(error?: Error): void {
        this.stopPingRoutine();

        if (this.connectionState === 'connected' || this.connectionState === 'connecting') {
            this.setConnectionState('disconnected');
            this.callbacks.onDisconnect?.(error);
        }

        // Rechazar todos los ACKs y respuestas pendientes
        this.rejectAllPending(error || new Error('Conexión cerrada'));

        // Intentar reconexión si está configurada
        if (this.reconnectAttempts < this.config.maxReconnectAttempts) {
            this.scheduleReconnect();
        } else {
            this.setConnectionState('failed');
            this.log('Máximo número de intentos de reconexión alcanzado');
        }
    }

    private scheduleReconnect(): void {
        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
        }

        this.setConnectionState('reconnecting');
        this.reconnectAttempts++;

        this.log(`Programando reconexión (intento ${this.reconnectAttempts}/${this.config.maxReconnectAttempts}) en ${this.config.reconnectInterval}ms`);

        this.reconnectTimer = setTimeout(() => {
            this.connect().catch((error) => {
                this.log('Error en reconexión', error);
            });
        }, this.config.reconnectInterval);
    }

    private rejectAllPending(error: Error): void {
        this.pendingAcks.forEach((pending) => pending.reject(error));
        this.pendingAcks.clear();

        this.pendingResponses.forEach((pending) => pending.reject(error));
        this.pendingResponses.clear();
    }

    private startPingRoutine(): void {
        this.stopPingRoutine();
        this.pingTimer = window.setInterval(() => {
            // Enviar un ping mediante un mensaje genérico
            if (this.ws?.readyState === WebSocket.OPEN) {
                try {
                    this.sendMessage({
                        type: 'generic_request',
                        payload: {
                            action: 'ping',
                            timestamp: Date.now()
                        }
                    });
                } catch (error) {
                    this.log('Error enviando ping', error);
                }
            }
        }, this.config.pingInterval);
    }

    private stopPingRoutine(): void {
        if (this.pingTimer) {
            clearInterval(this.pingTimer);
            this.pingTimer = null;
        }
    }

    private startCleanupRoutine(): void {
        this.cleanupTimer = setInterval(() => {
            const now = Date.now();

            // Limpiar ACKs expirados
            this.pendingAcks.forEach((pending, pid) => {
                if (now - pending.timestamp > this.config.ackTimeout) {
                    this.pendingAcks.delete(pid);
                    pending.reject(new Error(`Timeout esperando ACK para PID: ${pid}`));
                    this.log(`ACK timeout para PID: ${pid}`);
                }
            });

            // Limpiar respuestas expiradas
            this.pendingResponses.forEach((pending, pid) => {
                if (now - pending.timestamp > this.config.requestTimeout) {
                    this.pendingResponses.delete(pid);
                    pending.reject(new Error(`Timeout esperando respuesta para PID: ${pid}`));
                    this.log(`Response timeout para PID: ${pid}`);
                }
            });
        }, Math.min(this.config.ackTimeout, this.config.requestTimeout) / 2);
    }

    public sendMessage(message: ClientToServerMessage): void {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            throw new Error('WebSocket no está conectado');
        }

        if (!message.pid) {
            message.pid = this.generatePID();
        }

        const messageStr = JSON.stringify(message);
        this.ws.send(messageStr);
        this.log(`Mensaje enviado - Tipo: ${message.type}, PID: ${message.pid}, TargetUserId: ${message.targetUserId}`);
    }

    public sendMessageWithAck(message: ClientToServerMessage): Promise<ServerToClientMessage> {
        return new Promise((resolve, reject) => {
            if (!message.pid) {
                message.pid = this.generatePID();
            }

            const pendingAck: PendingAck = {
                resolve,
                reject,
                timestamp: Date.now(),
                messageId: message.pid
            };

            this.pendingAcks.set(message.pid, pendingAck);

            try {
                this.sendMessage(message);
            } catch (error) {
                this.pendingAcks.delete(message.pid);
                reject(error);
            }
        });
    }

    public sendRequestAndWaitResponse(message: ClientToServerMessage): Promise<ServerToClientMessage> {
        return new Promise((resolve, reject) => {
            if (!message.pid) {
                message.pid = this.generatePID();
            }

            const pendingResponse: PendingResponse = {
                resolve,
                reject,
                timestamp: Date.now()
            };

            this.pendingResponses.set(message.pid, pendingResponse);

            try {
                this.sendMessage(message);
            } catch (error) {
                this.pendingResponses.delete(message.pid);
                reject(error);
            }
        });
    }

    public sendClientAck(acknowledgedPid: string, status: string = 'received', error?: string): void {
        const ackMessage: ClientToServerMessage = {
            pid: this.generatePID(),
            type: 'client_ack',
            payload: {
                acknowledgedPid,
                status,
                error
            } as AckPayload
        };

        this.sendMessage(ackMessage);
    }

    // === MÉTODOS GENÉRICOS ACTUALIZADOS ===

    public sendDataRequest(payload: DataRequestPayload, targetUserId?: number): Promise<ServerToClientMessage> {
        const message: ClientToServerMessage = {
            type: 'data_request',
            targetUserId,
            payload
        };

        return this.sendMessageWithAck(message);
    }

    public sendPresenceUpdate(status: string, targetUserId?: number): void {
        const message: ClientToServerMessage = {
            type: 'presence_update',
            payload: { status, targetUserId } as PresenceUpdatePayload
        };

        this.sendMessage(message);
    }

    public sendGenericRequest(payload: any): Promise<ServerToClientMessage> {
        const message: ClientToServerMessage = {
            type: 'generic_request',
            payload
        };

        return this.sendRequestAndWaitResponse(message);
    }

    // === MÉTODOS DE CONVENIENCIA PARA CASOS DE USO ESPECÍFICOS ===

    /**
     * Envía un mensaje de chat privado usando el nuevo protocolo genérico
     */
    public sendChatMessage(text: string, targetUserId?: number): Promise<ServerToClientMessage> {
        const chatPayload: ChatMessagePayload = {
            action: 'send_message',
            resource: 'chat',
            data: {
                text,
                timestamp: new Date().toISOString()
            }
        };

        return this.sendDataRequest(chatPayload, targetUserId);
    }

    /**
     * Sube un chunk de archivo usando el nuevo protocolo genérico
     */
    public uploadFileChunk(
        fileName: string,
        fileSize: number,
        mimeType: string,
        chunkNum: number,
        totalChunks: number,
        fileId: string,
        data: string
    ): Promise<ServerToClientMessage> {
        const filePayload: FileUploadPayload = {
            action: 'upload_chunk',
            resource: 'file',
            data: {
                fileName,
                fileSize,
                mimeType,
                chunkNum,
                totalChunks,
                fileId,
                data
            }
        };

        return this.sendDataRequest(filePayload);
    }

    /**
     * Envía una notificación usando el nuevo protocolo genérico
     */
    public sendNotification(
        type: string,
        title: string,
        message: string,
        urgent: boolean = false,
        targetUsers?: number[]
    ): Promise<ServerToClientMessage> {
        const notificationPayload: NotificationPayload = {
            action: 'send',
            resource: 'notification',
            data: {
                type,
                title,
                message,
                urgent,
                targetUsers
            }
        };

        return this.sendDataRequest(notificationPayload);
    }

    /**
     * Solicita datos genéricos con action y resource específicos
     */
    public requestData(action: string, resource: string, data?: Record<string, any>): Promise<ServerToClientMessage> {
        const dataPayload: DataRequestPayload = {
            action,
            resource,
            data: data || {}
        };

        return this.sendDataRequest(dataPayload);
    }

    /**
     * Envía un mensaje peer-to-peer genérico
     */
    public sendPeerMessage(targetUserId: number, payload: any): Promise<ServerToClientMessage> {
        const message: ClientToServerMessage = {
            type: 'data_request',
            targetUserId,
            payload
        };

        return this.sendMessageWithAck(message);
    }

    public disconnect(): void {
        this.log('Desconectando WebSocket...');

        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
            this.reconnectTimer = null;
        }

        this.stopPingRoutine();

        if (this.ws) {
            this.ws.close(1000, 'Cliente desconectando');
            this.ws = null;
        }

        this.setConnectionState('disconnected');
        this.rejectAllPending(new Error('Cliente desconectado'));
    }

    public destroy(): void {
        this.disconnect();

        if (this.cleanupTimer) {
            clearInterval(this.cleanupTimer);
            this.cleanupTimer = null;
        }
    }

    public isConnected(): boolean {
        return this.ws?.readyState === WebSocket.OPEN;
    }

    public getPendingAcksCount(): number {
        return this.pendingAcks.size;
    }

    public getPendingResponsesCount(): number {
        return this.pendingResponses.size;
    }
}

// Función helper para crear una instancia del cliente
export function createWebSocketClient<TUserData = any>(
    config: WSClientConfig,
    callbacks?: WSClientCallbacks<TUserData>
): WebSocketClient<TUserData> {
    return new WebSocketClient<TUserData>(config, callbacks);
} 