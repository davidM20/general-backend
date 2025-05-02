let socket: WebSocket | null = null;
let messageHandler: ((event: MessageEvent) => void) | null = null;
let openHandler: (() => void) | null = null;
let closeHandler: (() => void) | null = null;
let errorHandler: ((event: Event) => void) | null = null;

// Asume que el proxy inverso está corriendo en localhost:8080
const WS_BASE_URL = 'ws://localhost:8080/ws';

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
        console.log('WebSocket message received:', event.data);
        try {
            const parsedData = JSON.parse(event.data);
            
            // Aquí podrías tener lógica para procesar diferentes tipos de mensajes
            messageHandler?.({ ...event, data: parsedData }); // Pasar datos parseados
        } catch (e) {
            console.error('Failed to parse WebSocket message:', e);
            // Podrías pasar el evento original si falla el parseo
            messageHandler?.(event);
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

// export { WebSocket }; // No es necesario re-exportar, es nativa del navegador 