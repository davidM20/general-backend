// Ejemplo de uso de la librería WebSocket Cliente actualizada
// Compatible con el paquete customws genérico de Go

import {
    WebSocketClient,
    createWebSocketClient,
    WSClientConfig,
    WSClientCallbacks,
    DataRequestPayload,
    ChatMessagePayload,
    FileUploadPayload,
    NotificationPayload
} from './websocket-client';

// 1. DEFINIR TIPO DE UserData (debe coincidir con el servidor Go)
interface MyUserData {
    userId: number;
    username: string;
    email: string;
    roles: string[];
    lastSeen: string;
    isActive: boolean;
    workspace: string;
}

// 2. CONFIGURAR CLIENTE CON AUTENTICACIÓN
const config: WSClientConfig = {
    url: 'ws://localhost:8082/ws',
    authHeaders: {
        'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
    },
    reconnectInterval: 3000,
    maxReconnectAttempts: 5,
    ackTimeout: 5000,
    requestTimeout: 10000,
    pingInterval: 30000,
    enableDebugLogs: true
};

// 3. CONFIGURAR CALLBACKS
const callbacks: WSClientCallbacks<MyUserData> = {
    onConnect: (userData) => {
        console.log('🔌 Conectado al servidor WebSocket!', userData);
    },

    onDisconnect: (error) => {
        console.log('❌ Desconectado del servidor', error?.message);
    },

    onConnectionStateChange: (state) => {
        console.log(`🔄 Estado de conexión: ${state}`);
    },

    // Manejar eventos de datos genéricos
    onDataEvent: (message) => {
        console.log('📦 Evento de datos recibido:', message);

        // Identificar el tipo de datos por payload
        const payload = message.payload as any;

        if (payload.event === 'new_message') {
            console.log('💬 Nuevo mensaje de chat:', payload.data);
        } else if (payload.event === 'file_uploaded') {
            console.log('📁 Archivo subido:', payload.data);
        } else if (payload.event === 'notification') {
            console.log('🔔 Nueva notificación:', payload.data);
        }
    },

    // Manejar eventos de presencia
    onPresenceEvent: (message) => {
        console.log('👤 Actualización de presencia:', message);
        const payload = message.payload as any;
        console.log(`Usuario ${payload.username} está ${payload.status}`);
    },

    // Manejar errores
    onErrorNotification: (message) => {
        console.error('🚨 Error del servidor:', message.error);
    },

    // Callback general para todos los mensajes
    onMessage: (message) => {
        console.log('📨 Mensaje recibido:', message.type, message);
    }
};

// 4. CREAR CLIENTE
const client = createWebSocketClient<MyUserData>(config, callbacks);

// 5. FUNCIONES DE EJEMPLO

async function connectAndDemo() {
    try {
        // Conectar
        await client.connect();
        console.log('✅ Conexión establecida');

        // Demo de diferentes tipos de comunicación
        await demoChat();
        await demoFileUpload();
        await demoNotifications();
        await demoPeerToPeer();
        await demoGenericRequests();

    } catch (error) {
        console.error('❌ Error conectando:', error);
    }
}

// DEMO: Chat usando protocolo genérico
async function demoChat() {
    console.log('\n🗣️ === DEMO CHAT ===');

    try {
        // Enviar mensaje de chat usando el método de conveniencia
        const response = await client.sendChatMessage('¡Hola mundo!');
        console.log('Chat enviado con ACK:', response);

        // Enviar mensaje de chat P2P a usuario específico
        const p2pResponse = await client.sendChatMessage('Mensaje privado', 456);
        console.log('Mensaje P2P enviado:', p2pResponse);

        // Solicitar historial de chat usando data request genérico
        const historyPayload: ChatMessagePayload = {
            action: 'get_history',
            resource: 'chat',
            data: {
                chatId: 'general',
            }
        };

        const historyResponse = await client.sendDataRequest(historyPayload);
        console.log('Historial de chat:', historyResponse);

    } catch (error) {
        console.error('Error en demo chat:', error);
    }
}

// DEMO: Upload de archivos por chunks
async function demoFileUpload() {
    console.log('\n📁 === DEMO FILE UPLOAD ===');

    try {
        const fileData = 'SGVsbG8gV29ybGQh'; // "Hello World!" en base64
        const fileId = `file_${Date.now()}`;

        // Subir chunk único
        const uploadResponse = await client.uploadFileChunk(
            'test.txt',
            13, // tamaño en bytes
            'text/plain',
            1, // chunk número 1
            1, // total de chunks
            fileId,
            fileData
        );

        console.log('Archivo subido:', uploadResponse);

        // Completar upload
        const completePayload: FileUploadPayload = {
            action: 'complete_upload',
            resource: 'file',
            data: {
                fileName: 'test.txt',
                fileSize: 13,
                mimeType: 'text/plain',
                chunkNum: 1,
                totalChunks: 1,
                fileId,
                data: ''
            }
        };

        const completeResponse = await client.sendDataRequest(completePayload);
        console.log('Upload completado:', completeResponse);

    } catch (error) {
        console.error('Error en demo file upload:', error);
    }
}

// DEMO: Sistema de notificaciones
async function demoNotifications() {
    console.log('\n🔔 === DEMO NOTIFICATIONS ===');

    try {
        // Enviar notificación general
        const notifResponse = await client.sendNotification(
            'info',
            'Sistema actualizado',
            'El sistema ha sido actualizado a la versión 2.0',
            false
        );
        console.log('Notificación enviada:', notifResponse);

        // Enviar notificación urgente a usuarios específicos
        const urgentResponse = await client.sendNotification(
            'alert',
            'Mantenimiento programado',
            'El sistema entrará en mantenimiento en 5 minutos',
            true,
            [123, 456, 789] // IDs de usuarios específicos
        );
        console.log('Notificación urgente enviada:', urgentResponse);

        // Obtener notificaciones pendientes
        const pendingPayload: NotificationPayload = {
            action: 'get_pending',
            resource: 'notification',
            data: {}
        };

        const pendingResponse = await client.sendDataRequest(pendingPayload);
        console.log('Notificaciones pendientes:', pendingResponse);

    } catch (error) {
        console.error('Error en demo notifications:', error);
    }
}

// DEMO: Comunicación Peer-to-Peer
async function demoPeerToPeer() {
    console.log('\n👥 === DEMO PEER-TO-PEER ===');

    try {
        const targetUserId = 456;

        // Mensaje P2P genérico
        const p2pResponse = await client.sendPeerMessage(targetUserId, {
            type: 'collaboration_invite',
            data: {
                projectId: 'proj_123',
                projectName: 'Proyecto Importante',
                role: 'editor',
                message: '¿Te gustaría colaborar en este proyecto?'
            }
        });
        console.log('Mensaje P2P enviado:', p2pResponse);

        // Video call P2P
        const videoCallPayload = {
            action: 'video_call_request',
            resource: 'communication',
            data: {
                callId: `call_${Date.now()}`,
                callerName: 'Mi Usuario',
                sdpOffer: 'v=0\r\no=...' // SDP de WebRTC
            }
        };

        const callResponse = await client.sendDataRequest(videoCallPayload, targetUserId);
        console.log('Llamada iniciada:', callResponse);

    } catch (error) {
        console.error('Error en demo P2P:', error);
    }
}

// DEMO: Solicitudes genéricas
async function demoGenericRequests() {
    console.log('\n🔧 === DEMO GENERIC REQUESTS ===');

    try {
        // Solicitud de datos de perfil
        const profileData = await client.requestData('get', 'user_profile', {
            userId: 123,
            fields: ['username', 'email', 'roles', 'lastSeen']
        });
        console.log('Datos de perfil:', profileData);

        // Búsqueda genérica
        const searchData = await client.requestData('search', 'documents', {
            query: 'importante',
            filters: {
                dateRange: 'last_week',
                type: 'pdf'
            },
            limit: 10
        });
        console.log('Resultados de búsqueda:', searchData);

        // Configuración de aplicación
        const configData = await client.requestData('get', 'app_config', {
            section: 'websocket',
            includeDefaults: true
        });
        console.log('Configuración:', configData);

        // Solicitud con respuesta esperada específica
        const customResponse = await client.sendGenericRequest({
            operation: 'custom_operation',
            parameters: {
                param1: 'value1',
                param2: 42,
                param3: true
            }
        });
        console.log('Respuesta personalizada:', customResponse);

    } catch (error) {
        console.error('Error en demo generic requests:', error);
    }
}

// DEMO: Actualización de presencia
function demoPresence() {
    console.log('\n👤 === DEMO PRESENCE ===');

    // Actualizar estado general
    client.sendPresenceUpdate('online');

    // Actualizar estado para usuario específico (ej. "escribiendo" en chat)
    client.sendPresenceUpdate('typing', 456);

    // Simular cambios de estado
    setTimeout(() => {
        client.sendPresenceUpdate('away');
    }, 10000);

    setTimeout(() => {
        client.sendPresenceUpdate('offline');
    }, 20000);
}

// 6. MANEJO DE EVENTOS DE CICLO DE VIDA

window.addEventListener('beforeunload', () => {
    console.log('🔌 Cerrando conexión antes de salir...');
    client.disconnect();
});

// Reconexión manual
document.getElementById('reconnect-btn')?.addEventListener('click', async () => {
    if (!client.isConnected()) {
        try {
            await client.connect();
            console.log('✅ Reconectado manualmente');
        } catch (error) {
            console.error('❌ Error en reconexión manual:', error);
        }
    }
});

// Estado de la conexión
setInterval(() => {
    const state = client.getConnectionState();
    const pendingAcks = client.getPendingAcksCount();
    const pendingResponses = client.getPendingResponsesCount();

    document.getElementById('status')!.textContent =
        `Estado: ${state} | ACKs pendientes: ${pendingAcks} | Respuestas pendientes: ${pendingResponses}`;
}, 1000);

// 7. INICIAR DEMO
console.log('🚀 Iniciando demo del cliente WebSocket...');
connectAndDemo();
demoPresence();

export { client }; 