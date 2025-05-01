import React, { useState, useEffect, useCallback } from 'react';
import {
    connectWebSocket,
    disconnectWebSocket,
    // sendWebSocketMessage, // Ya no enviaremos mensajes desde aquí directamente
    setWebSocketMessageHandler,
    setWebSocketOpenHandler,
    setWebSocketCloseHandler,
    setWebSocketErrorHandler,
    getWebSocketState,
} from '../services/websocket';
import { useAuth } from '../contexts/AuthContext';

const TestWebSocketConnect: React.FC = () => {
    const { token } = useAuth(); // Obtener token del contexto
    const [isConnected, setIsConnected] = useState<boolean>(false);
    // const [messageToSend, setMessageToSend] = useState<string>('...'); // Eliminado
    const [receivedMessages, setReceivedMessages] = useState<any[]>([]);
    const [error, setError] = useState<string | null>(null);

    const updateConnectionStatus = useCallback(() => {
        const state = getWebSocketState();
        setIsConnected(state === 1); // WebSocket.OPEN es 1
    }, []);

    useEffect(() => {
        // Configurar manejadores globales
        setWebSocketOpenHandler(() => {
            setError(null);
            updateConnectionStatus();
            setReceivedMessages(prev => [
                { type: 'status', data: `Connected! (${new Date().toLocaleTimeString()})` },
                ...prev
            ]);
        });

        setWebSocketMessageHandler((event: MessageEvent) => {
             // Insertar al principio para ver los más recientes primero
            setReceivedMessages(prev => [
                { type: 'message', data: event.data, time: new Date().toLocaleTimeString() },
                ...prev
            ]);
        });

        setWebSocketErrorHandler((event: Event) => {
            console.error('WebSocket Error Handler:', event);
            setError('WebSocket connection error.');
            updateConnectionStatus();
            setReceivedMessages(prev => [
                 { type: 'status', data: `Error! (${new Date().toLocaleTimeString()})` },
                 ...prev
            ]);
        });

        setWebSocketCloseHandler(() => {
            //setError('WebSocket disconnected.'); // No mostrar error, es normal desconectarse
            updateConnectionStatus();
            setReceivedMessages(prev => [
                { type: 'status', data: `Disconnected! (${new Date().toLocaleTimeString()})` },
                ...prev
             ]);
        });

        updateConnectionStatus(); // Verificar estado inicial

        // Limpiar manejadores al desmontar esta página específica?
        // Probablemente no, queremos que los manejadores persistan mientras la app esté viva.
        // return () => { ... };

    }, [updateConnectionStatus]);

    const handleConnect = () => {
        if (!token) {
            setError('Cannot connect: No JWT Token found. Please login first.');
            return;
        }
        setError(null);
        // No limpiar mensajes al reconectar, mantener historial
        // setReceivedMessages([]);
        connectWebSocket(token);
    };

    const handleDisconnect = () => {
        disconnectWebSocket();
    };

    // const handleSendMessage = () => { ... }; // Eliminado

    return (
        <div>
            <h2>WebSocket Connection / Status</h2>
            <p>Usa esta sección para conectar/desconectar el WebSocket usando el token JWT obtenido del login.</p>
            <p>El token actual es: <code>{token ? `${token.substring(0, 15)}...` : 'N/A'}</code></p>
            <div>
                {/* Ya no necesitamos input para token aquí */}
                <button onClick={handleConnect} disabled={!token || isConnected} style={{ marginLeft: '0rem' }}>
                    Connect WS
                </button>
                <button onClick={handleDisconnect} disabled={!isConnected} style={{ marginLeft: '0.5rem' }}>
                    Disconnect WS
                </button>
                <span style={{ marginLeft: '1rem' }}>Status: {isConnected ? 'Connected' : 'Disconnected'}</span>
            </div>

            {/* Ya no mostramos el área para enviar mensajes */}
            {/* {isConnected && (...)} */}

            {error && <div style={{ color: 'red', marginTop: '1rem' }}>Error: {error}</div>}

            <div style={{ marginTop: '1rem' }}>
                <h3>Received Messages (Global):</h3>
                <p>Todos los mensajes recibidos por el WebSocket se mostrarán aquí.</p>
                 <button onClick={() => setReceivedMessages([])} style={{marginBottom: '0.5rem'}} disabled={receivedMessages.length === 0}>
                    Clear Log
                </button>
                <div style={{ height: '400px', overflowY: 'scroll', border: '1px solid #ccc', padding: '0.5rem', background: '#f9f9f9' }}>
                    {receivedMessages.map((msg, index) => (
                        <div key={index} style={{ borderBottom: '1px dashed #eee', marginBottom:'0.5rem', paddingBottom:'0.5rem'}}>
                           <span style={{fontSize: '0.8em', color: '#777'}}>{msg.time || new Date().toLocaleTimeString()} - {msg.type === 'status' ? 'Status' : 'Message'}</span>
                           <pre style={{ margin: 0, background: msg.type === 'status' ? '#e8f0fe' : '#fff' }}>
                                {JSON.stringify(msg.data, null, 2)}
                            </pre>
                        </div>
                    ))}
                    {receivedMessages.length === 0 && <p>No messages received yet.</p>}
                </div>
            </div>
        </div>
    );
};

export default TestWebSocketConnect; 