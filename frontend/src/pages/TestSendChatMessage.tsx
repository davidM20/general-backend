import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import {
    sendWebSocketMessage,
    getWebSocketState,
    // Importar constantes de estado si se exportan desde websocket.ts
    // o usar los números directamente: 0: CONNECTING, 1: OPEN, 2: CLOSING, 3: CLOSED
} from '../services/websocket';
import { toast } from 'react-toastify'; // Importar toast

const TestSendChatMessage: React.FC = () => {
    const { token } = useAuth();
    const [toUserId, setToUserId] = useState<string>('');
    const [text, setText] = useState<string>('');
    const [isConnected, setIsConnected] = useState<boolean>(false);

    // Verificar el estado de la conexión periódicamente o al cargar
    useEffect(() => {
        const checkConnection = () => {
            setIsConnected(getWebSocketState() === 1); // 1 = WebSocket.OPEN
        };
        checkConnection(); // Comprobar al montar
        const intervalId = setInterval(checkConnection, 2000); // Comprobar cada 2 segundos
        return () => clearInterval(intervalId); // Limpiar intervalo al desmontar
    }, []);

    const handleSubmit = (event: React.FormEvent) => {
        event.preventDefault();
        if (!isConnected) {
            toast.error('WebSocket is not connected. Please connect first.');
            return;
        }
        if (!toUserId || !text) {
            toast.error('Both User ID and Text are required.');
            return;
        }

        const userIdNum = parseInt(toUserId, 10);
        if (isNaN(userIdNum)) {
            toast.error('Invalid User ID. Please enter a number.');
            return;
        }

        const message = {
            type: "chat_message",
            payload: {
                toUserId: userIdNum,
                text: text
            }
        };

        try {
            sendWebSocketMessage(message);
            toast.info(`Message sent to user ${toUserId}.`); // Toast informativo
            setToUserId('');
            setText('');
        } catch (e) {
            console.error("Error sending message:", e);
            toast.error('Failed to send message. Check console.');
        }
    };

    return (
        <div>
            <h2>Send Chat Message (WS)</h2>
            <p>Usa este formulario para enviar un mensaje de chat a otro usuario a través de WebSocket.</p>
            <p>Estado de WebSocket: <strong>{isConnected ? 'Conectado' : 'Desconectado'}</strong></p>
            {!token && <p style={{color: 'orange'}}>Advertencia: No hay token JWT. Necesitarás uno para conectarte.</p>}
            {!isConnected && token && <p style={{color: 'orange'}}>Advertencia: WebSocket desconectado. Ve a la página 'Connect (WS)' para conectar.</p>}

            <form onSubmit={handleSubmit} style={{ maxWidth: '500px', marginTop: '1rem' }}>
                 <div style={{ marginBottom: '1rem' }}>
                    <label htmlFor="toUserId">To User ID:</label>
                    <input
                        type="number"
                        id="toUserId"
                        value={toUserId}
                        onChange={(e) => setToUserId(e.target.value)}
                        required
                        style={{ width: '100%', marginTop: '0.3rem' }}
                        disabled={!isConnected}
                    />
                </div>
                 <div style={{ marginBottom: '1rem' }}>
                    <label htmlFor="text">Message Text:</label>
                    <textarea
                        id="text"
                        value={text}
                        onChange={(e) => setText(e.target.value)}
                        required
                        rows={3}
                        style={{ width: '100%', marginTop: '0.3rem' }}
                        disabled={!isConnected}
                    />
                </div>
                <button type="submit" disabled={!isConnected} style={{ width: '100%' }}>
                    Send Chat Message
                </button>
            </form>
        </div>
    );
};

export default TestSendChatMessage; 