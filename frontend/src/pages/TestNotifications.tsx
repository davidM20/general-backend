import React, { useState, useEffect, useCallback } from 'react';
import {
    addWebSocketListener,
    removeWebSocketListener,
    sendWebSocketMessage,
    getWebSocketState
} from '../services/websocket';
import {
    MessageTypeGetNotifications, // Constante para la solicitud
    MessageTypeNotificationsResponse, // Constante para la respuesta
    NotificationsResponsePayload, // Payload de la respuesta
    NotificationInfo, // Tipo de dato individual
    MessageTypeError // Para errores
} from '../types/websocket';
import { toast } from 'react-toastify';

// Tipo para el mensaje parseado (puede eliminarse si no se usa fuera del useEffect eliminado)
/*
interface ParsedWebSocketMessage {
    type: string;
    payload: any;
    error?: string;
}
*/

const TestNotifications: React.FC = () => {
   
    // Eliminar el estado isConnected si ya no se usa
    // const [isConnected, setIsConnected] = useState<boolean>(false);
    const [notificationIdsToMark, setNotificationIdsToMark] = useState<string>('');
    const [notifications, setNotifications] = useState<NotificationInfo[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // Callback para manejar la respuesta de notificaciones
    const handleNotificationsResponse = useCallback((payload: NotificationInfo[], errorMsg?: string) => {
        console.log(`handleNotificationsResponse (listening on ${MessageTypeGetNotifications}) called:`, payload, "Error:", errorMsg);
        setIsLoading(false);
        if (errorMsg) {
            const message = `Failed to get notifications: ${errorMsg}`;
            setError(message);
            toast.error(message);
            setNotifications([]);
        } else if (Array.isArray(payload)) {
            setNotifications(payload);
            setError(null);
            toast.success(`${payload.length} notifications received.`);
        } else {
             const message = "Received invalid data format for notifications.";
             setError(message);
             toast.warn(message);
             setNotifications([]);
        }
    }, []);

    // Callback para manejar errores generales de WS
    const handleErrorResponse = useCallback((payload: { error: string } | any, errorMsg?: string) => {
        setIsLoading(false);
        const message = errorMsg || payload?.error || "An unknown WebSocket error occurred.";
        console.error("handleErrorResponse (Notifications) called:", message);
        setError(message);
        toast.error(message);
        setNotifications([]); // Limpiar datos en caso de error
    }, []);

    // Efecto para registrar/desregistrar listeners
    useEffect(() => {
        console.log("Registering WS listeners for notifications page...");
        addWebSocketListener(MessageTypeGetNotifications, handleNotificationsResponse);
        addWebSocketListener(MessageTypeError, handleErrorResponse);

        // Solicitar notificaciones al montar si WS está conectado
         if (getWebSocketState() === WebSocket.OPEN) {
              fetchNotifications();
         } else {
              setError("WebSocket not connected on mount.");
         }

        // Función de limpieza al desmontar
        return () => {
            console.log("Unregistering WS listeners for notifications page...");
            removeWebSocketListener(MessageTypeGetNotifications, handleNotificationsResponse);
            removeWebSocketListener(MessageTypeError, handleErrorResponse);
        };
    }, [handleNotificationsResponse, handleErrorResponse]);

    // Función para solicitar notificaciones
    const fetchNotifications = () => {
        const wsState = getWebSocketState();
        if (wsState !== WebSocket.OPEN) {
            toast.error("WebSocket is not connected. Cannot fetch notifications.");
            setError("WebSocket is not connected.");
            return;
        }
        setIsLoading(true);
        setError(null);
        setNotifications([]); // Limpiar datos anteriores
        console.log(`Sending '${MessageTypeGetNotifications}' request via WebSocket...`);
        sendWebSocketMessage({
            type: MessageTypeGetNotifications,
            payload: {} // Sin payload necesario para get-notifications
        });
        toast.info("Request sent for notifications.");
    };

    const handleMarkNotificationsRead = () => {
        if (getWebSocketState() !== WebSocket.OPEN) {
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
            <h2>Notifications (WebSocket Test)</h2>
            <button onClick={fetchNotifications} disabled={isLoading || getWebSocketState() !== WebSocket.OPEN}>
                {isLoading ? 'Loading...' : 'Fetch Notifications'}
            </button>
            {getWebSocketState() !== WebSocket.OPEN && <p style={{color: 'orange', display: 'inline', marginLeft: '10px'}}>WebSocket Disconnected</p>}

            {error && <p style={{ color: 'red', marginTop: '10px' }}>Error: {error}</p>}

            <div style={{ marginTop: '20px', background: '#f9f9f9', padding: '15px', borderRadius: '5px', border: '1px solid #eee', minHeight: '200px' }}>
                <h3>Notifications Received:</h3>
                {notifications.length > 0 ? (
                    <pre style={{ whiteSpace: 'pre-wrap', wordWrap: 'break-word' }}>
                        {JSON.stringify(notifications, null, 2)}
                    </pre>
                ) : (
                     !isLoading && !error && getWebSocketState() === WebSocket.OPEN && <p>No notifications received or list is empty.</p>
                )}
                 {isLoading && <p>Loading notifications...</p>}
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
                        disabled={getWebSocketState() !== WebSocket.OPEN}
                         style={{ width: '100%', maxWidth: '400px', marginTop: '0.3rem', display:'block' }}
                    />
                </div>
                 <button
                    onClick={handleMarkNotificationsRead}
                    disabled={getWebSocketState() !== WebSocket.OPEN || !notificationIdsToMark.trim()}
                >
                    Mark as Read (Disabled)
                </button>
            </div>
        </div>
    );
};

export default TestNotifications; 