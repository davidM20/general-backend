import {
    SearchPayload,
    SearchResponsePayload,
    MessageTypeSearch,
    MessageTypeSearchResponse,
    IncomingMessage,
    OutgoingMessage,
    TypedSearchResult,
    UserSearchResult,
    EnterpriseSearchResult
} from '../types/websocket';
import { toast } from 'react-toastify';

let socket: WebSocket | null = null;
let messageHandler: ((event: MessageEvent) => void) | null = null;
let openHandler: (() => void) | null = null;
let closeHandler: (() => void) | null = null;
let errorHandler: ((event: Event) => void) | null = null;

// Asume que el proxy inverso está corriendo en localhost:8080
const WS_BASE_URL = 'ws://localhost:8080/ws';

// Añadir un tipo para los listeners
type MessageHandler<T = any> = (payload: T, error?: string) => void;
const listeners = new Map<string, MessageHandler[]>();

// Función para registrar listeners (modificada o añadida)
export const addWebSocketListener = <T>(messageType: string, handler: MessageHandler<T>) => {
    if (!listeners.has(messageType)) {
        listeners.set(messageType, []);
    }
    listeners.get(messageType)?.push(handler as MessageHandler); // Cast a MessageHandler genérico
};

// Función para eliminar listeners (modificada o añadida)
export const removeWebSocketListener = <T>(messageType: string, handler: MessageHandler<T>) => {
    const messageListeners = listeners.get(messageType);
    if (messageListeners) {
        listeners.set(
            messageType,
            messageListeners.filter((h) => h !== handler as MessageHandler) // Cast a MessageHandler genérico
        );
    }
};

export const connectWebSocket = (token: string) => {
    if (socket && socket.readyState === WebSocket.OPEN) {
        console.log('WebSocket is already connected.');
        return;
    }

    if (!token) {
        console.error('Token is required to connect WebSocket');
        return;
    }

    // Añadir el token como parámetro de URL
    const url = `${WS_BASE_URL}?token=${encodeURIComponent(token)}`;
    console.log('Connecting WebSocket to:', url);

    socket = new WebSocket(url);

    socket.onopen = () => {
        console.log('WebSocket connected');
        openHandler?.();
    };

    socket.onmessage = (event) => {
        try {
            const message: IncomingMessage = JSON.parse(event.data);
            console.log('WebSocket message received:', message);

            // Log para depuración de datos crudos
            console.log('Raw WS Data:', event.data);

            const handlers = listeners.get(message.type);
            if (handlers) {
                handlers.forEach((handler) => {
                     // Distinguir resultados antes de llamar al handler de search_response
                     if (message.type === MessageTypeSearchResponse) {
                        const payload = message.payload as SearchResponsePayload;
                        const typedResults: TypedSearchResult[] = payload.results.map(result => {
                            // Asumiendo que podemos distinguir por una propiedad única,
                            // por ejemplo, 'rif' solo existe en EnterpriseSearchResult
                            if ('rif' in result) {
                                 return { ...(result as EnterpriseSearchResult), resultType: 'enterprise' };
                            } else {
                                 return { ...(result as UserSearchResult), resultType: 'user' };
                            }
                        });
                        const typedPayload = { ...payload, results: typedResults };
                        handler(typedPayload, message.error);
                    } else {
                        handler(message.payload, message.error);
                    }
                });
            } else {
                console.warn(`No listener registered for message type: ${message.type}`);
                // Opcional: mostrar un toast genérico si no hay handler específico
                if (message.type === 'success' && message.payload?.message) {
                    toast.success(message.payload.message);
                } else if (message.error) {
                     toast.error(`Error: ${message.error}`);
                }
            }

        } catch (error) {
            console.error('Failed to parse WebSocket message or execute handler:', error);
            toast.error('Received invalid message format from server.');
        }
    };

    socket.onerror = (event) => {
        console.error('WebSocket error:', event);
        errorHandler?.(event);
    };

    socket.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        socket = null; // Limpiar la referencia
        closeHandler?.();
    };
};

export const disconnectWebSocket = () => {
    if (socket) {
        socket.close();
        console.log('WebSocket disconnecting...');
    } else {
        console.log('WebSocket is not connected.');
    }
};

export const sendWebSocketMessage = (message: any) => {
    if (socket && socket.readyState === WebSocket.OPEN) {
        try {
            const jsonMessage = JSON.stringify(message);
            socket.send(jsonMessage);
            console.log('WebSocket message sent:', jsonMessage);
        } catch (e) {
            console.error('Failed to stringify message:', e);
        }
    } else {
        console.error('WebSocket is not connected or not open.');
    }
};

export const setWebSocketMessageHandler = (handler: (event: MessageEvent) => void) => {
    messageHandler = handler;
};

export const setWebSocketOpenHandler = (handler: () => void) => {
    openHandler = handler;
};

export const setWebSocketCloseHandler = (handler: () => void) => {
    closeHandler = handler;
};

export const setWebSocketErrorHandler = (handler: (event: Event) => void) => {
    errorHandler = handler;
};

export const getWebSocketState = (): number | null => {
    return socket?.readyState ?? null;
}

// --- Nueva Función para Enviar Búsqueda ---
export const sendSearchRequest = (query: string, entityType: 'user' | 'enterprise' | 'all', limit: number = 20, offset: number = 0) => {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
        console.error('WebSocket is not connected.');
        toast.error('WebSocket connection lost. Cannot send search request.');
        return;
    }

    const payload: SearchPayload = { query, entityType, limit, offset };
    const message: OutgoingMessage<SearchPayload> = {
        type: MessageTypeSearch,
        payload: payload,
    };

    try {
        const jsonMessage = JSON.stringify(message);
        console.log('Sending WebSocket message:', jsonMessage);
        socket.send(jsonMessage);
    } catch (error) {
        console.error('Failed to stringify or send WebSocket message:', error);
        toast.error('Failed to send search request.');
    }
};

// export { WebSocket }; // No es necesario re-exportar, es nativa del navegador 