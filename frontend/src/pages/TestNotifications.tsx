import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import {
    sendWebSocketMessage,
    getWebSocketState,
    setWebSocketMessageHandler,
} from '../services/websocket';
import { toast } from 'react-toastify';

// Interface para la estructura de Event (simplificada)
interface NotificationEvent {
    Id: number;
    Description: string;
    UserId: number;
    OtherUserId?: number | null;
    ProyectId?: number | null;
    CreateAt: string; // Asumiendo que llega como string ISO
    // IsRead no se incluye ya que se eliminó
}

// Tipo para el mensaje parseado
interface ParsedWebSocketMessage {
    type: string;
    payload: any;
    error?: string;
}

const TestNotifications: React.FC = () => {
    const { token } = useAuth();
    const [isConnected, setIsConnected] = useState<boolean>(false);
    const [notificationIdsToMark, setNotificationIdsToMark] = useState<string>('');

    // Estados para mostrar resultados
    const [notifications, setNotifications] = useState<NotificationEvent[] | null>(null);
    const [notificationError, setNotificationError] = useState<string | null>(null);

    useEffect(() => {
        const checkConnection = () => {
            setIsConnected(getWebSocketState() === 1);
        };
        checkConnection();
        const intervalId = setInterval(checkConnection, 2000);

        // Handler para mensajes de esta página
        const handleWsMessage = (event: MessageEvent) => {
            const parsedData = event.data as ParsedWebSocketMessage;

            if (parsedData.type === 'get-notifications') {
                setNotificationError(null);
                setNotifications(parsedData.payload as NotificationEvent[]);
                console.log('Received notifications:', parsedData.payload);
            } else if (parsedData.type === 'error' && parsedData.error?.includes('notification')) { 
                // Intentar detectar errores relacionados con notificaciones
                console.error('Notification error from backend:', parsedData.error);
                setNotificationError(parsedData.error || 'Unknown error fetching notifications');
                setNotifications(null);
            }
             // Ignorar otros tipos de mensajes
        };

        // Registrar handler
        setWebSocketMessageHandler(handleWsMessage);

        return () => {
            clearInterval(intervalId);
             // setWebSocketMessageHandler(null); // Opcional: desregistrar
        };
    }, []);

    const handleGetNotifications = () => {
        if (!isConnected) {
            toast.error('WebSocket is not connected. Please connect first.');
            return;
        }
        // Limpiar estado previo
        setNotifications(null);
        setNotificationError(null);

        const message = { type: 'get-notifications', payload: {} };
        try {
            sendWebSocketMessage(message);
            toast.info('Request sent: get-notifications.');
        } catch (e) {
            console.error('Error sending get-notifications request:', e);
            const errorMsg = e instanceof Error ? e.message : String(e);
            toast.error(`Failed to send get-notifications request: ${errorMsg}`);
            setNotificationError(`Frontend error sending request: ${errorMsg}`);
        }
    };

    const handleMarkNotificationsRead = () => {
        if (!isConnected) {
            toast.error('WebSocket is not connected. Please connect first.');
            return;
        }
        if (!notificationIdsToMark.trim()) {
            toast.error('Please enter Notification IDs (comma-separated) to mark as read.');
            return;
        }

        // Convertir string CSV a array de números
        const ids = notificationIdsToMark.split(',').map(id => parseInt(id.trim(), 10)).filter(id => !isNaN(id));

        if (ids.length === 0) {
             toast.error('Invalid Notification IDs entered. Please use comma-separated numbers.');
            return;
        }

        // NOTA: El backend ya no soporta marcar como leídas
        // Esta lógica fallará o será ignorada por el backend.
        // Debería eliminarse o adaptarse si se reimplementa la funcionalidad "IsRead".
        const message = {
            type: 'mark_notifications_read', 
            payload: { notificationIds: ids }
        };
        toast.warn('Functionality to mark notifications as read is currently disabled in backend.');

        /* // Comentado para evitar envío
        try {
            sendWebSocketMessage(message);
            toast.info(`Request sent: mark_notifications_read for IDs: ${ids.join(', ')}. Check WS Connect log.`);
             setNotificationIdsToMark(''); // Limpiar input
        } catch (e) {
            console.error('Error sending mark_notifications_read request:', e);
            toast.error('Failed to send mark_notifications_read request. Check console.');
        }
        */
    };

    return (
        <div>
            <h2>Notifications (WS)</h2>
            <p>Obtener notificaciones.</p> 
            <p>Estado de WebSocket: <strong>{isConnected ? 'Conectado' : 'Desconectado'}</strong></p>
            {!token && <p style={{color: 'orange'}}>Advertencia: No hay token JWT.</p>}
            {!isConnected && token && <p style={{color: 'orange'}}>Advertencia: WebSocket desconectado. Ve a 'Connect (WS)'.</p>}

            <div style={{ marginTop: '1rem' }}>
                <button
                    onClick={handleGetNotifications}
                    disabled={!isConnected}
                    style={{ marginRight: '1rem' }}
                >
                    Get My Notifications
                </button>
            </div>

            {/* Sección para mostrar resultados */}        
             <div style={{ marginTop: '1rem', border: '1px solid #eee', padding: '1rem', minHeight: '150px', background: '#f8f8f8' }}>
                <h3>Response:</h3>
                {notificationError && (
                    <div style={{ color: 'red', marginBottom: '1rem' }}>
                        <strong>Error:</strong> {notificationError}
                    </div>
                )}
                {notifications && (
                    <div>
                        <h4>Notifications ({notifications.length})</h4>
                        <pre style={{ maxHeight: '300px', overflowY: 'auto', background: 'white', padding: '0.5rem' }}>
                            {JSON.stringify(notifications, null, 2)}
                        </pre>
                    </div>
                )}
                 {!notificationError && !notifications && (
                     <p>Click the button above to request notifications.</p>
                )}
            </div>

            <hr style={{ margin: '2rem 0' }}/>

             {/* Sección Mark as Read - Deshabilitada funcionalmente */}
            <div>
                <h3>Mark Notifications as Read (Currently Disabled)</h3>
                 <p style={{fontStyle: 'italic', color: '#777'}}>La funcionalidad para marcar notificaciones como leídas se ha desactivado en el backend.</p>
                <div style={{ marginBottom: '0.5rem' }}>
                    <label htmlFor="notificationIds">Notification IDs (comma-separated):</label>
                    <input
                        type="text"
                        id="notificationIds"
                        value={notificationIdsToMark}
                        onChange={(e) => setNotificationIdsToMark(e.target.value)}
                        placeholder="e.g., 1, 5, 10"
                        disabled={!isConnected}
                         style={{ width: '100%', maxWidth: '400px', marginTop: '0.3rem', display:'block' }}
                    />
                </div>
                 <button
                    onClick={handleMarkNotificationsRead}
                    disabled={!isConnected || !notificationIdsToMark.trim()}
                >
                    Mark as Read (Disabled)
                </button>
            </div>
        </div>
    );
};

export default TestNotifications; 