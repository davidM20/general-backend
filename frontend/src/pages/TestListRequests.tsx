import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import {
    sendWebSocketMessage,
    getWebSocketState,
    setWebSocketMessageHandler,
} from '../services/websocket';
import { toast } from 'react-toastify';

// Define the *payload* list types
type ListPayloadType = 'contacts' | 'chats' | 'online_users';

// Interfaces para los datos esperados (copiadas/adaptadas de types.go)
interface ContactInfo {
    userId: number;
    userName: string;
    firstName?: string;
    lastName?: string;
    picture?: string;
    isOnline: boolean;
    chatId: string;
}

interface ChatInfo {
    chatId: string;
    otherUserId: number;
    otherUserName: string;
    otherPicture?: string;
    lastMessage?: string;
    lastMessageTs?: number;
    unreadCount?: number;
    isOtherOnline: boolean;
}

interface OnlineUserInfo {
    userId: number;
    userName: string;
}

// Tipo para el mensaje parseado que esperamos del handler
// Idealmente, esto debería venir de websocket.ts
interface ParsedWebSocketMessage {
    type: string;
    payload: any; // Ser más específico si es posible
    error?: string;
}

const TestListRequests: React.FC = () => {
    const { token } = useAuth();
    const [isConnected, setIsConnected] = useState<boolean>(false);

    // Estados para guardar las listas recibidas y errores
    const [contacts, setContacts] = useState<ContactInfo[] | null>(null);
    const [chats, setChats] = useState<ChatInfo[] | null>(null);
    const [onlineUsers, setOnlineUsers] = useState<OnlineUserInfo[] | null>(null);
    const [listError, setListError] = useState<string | null>(null);

    useEffect(() => {
        const checkConnection = () => {
            setIsConnected(getWebSocketState() === 1);
        };
        checkConnection();
        const intervalId = setInterval(checkConnection, 2000);

        // Definir el handler para mensajes WebSocket específicos de esta página
        const handleWsMessage = (event: MessageEvent) => {
            // event.data ya debería estar parseado por el servicio websocket.ts
            const parsedData = event.data as ParsedWebSocketMessage; 

            // Solo procesar mensajes de tipo 'list' que son respuestas a nuestras solicitudes
            if (parsedData.type === 'list') { 
                setListError(null); // Limpiar error previo
                const { listType, data } = parsedData.payload;
                console.log(`Received list response for type: ${listType}`, data);
                switch (listType) {
                    case 'contacts':
                        setContacts(data as ContactInfo[]);
                        setChats(null); // Limpiar otros estados
                        setOnlineUsers(null);
                        break;
                    case 'chats':
                        setChats(data as ChatInfo[]);
                        setContacts(null);
                        setOnlineUsers(null);
                        break;
                    case 'online_users':
                        setOnlineUsers(data as OnlineUserInfo[]);
                        setContacts(null);
                        setChats(null);
                        break;
                    default:
                        console.warn('Received unknown list type:', listType);
                        setListError(`Received unknown list type in payload: ${listType}`);
                        break;
                }
            } else if (parsedData.type === 'error' && parsedData.payload?.requestType === 'list') {
                 // Si el backend envía un error específico para la lista
                 console.error('List request error from backend:', parsedData.error || parsedData.payload?.message);
                 setListError(parsedData.error || parsedData.payload?.message || 'Unknown error fetching list');
                 setContacts(null);
                 setChats(null);
                 setOnlineUsers(null);
            }
        };

        // Registrar el handler
        setWebSocketMessageHandler(handleWsMessage);

        // Limpieza al desmontar el componente
        return () => {
            clearInterval(intervalId);
            // Quizás desregistrar el handler si es necesario para evitar 
            // que se ejecute cuando el componente no está montado.
            // setWebSocketMessageHandler(null); // O una función vacía
        };
    }, []); // Dependencias vacías para que se ejecute solo al montar/desmontar

    const handleSendListRequest = (listPayloadType: ListPayloadType) => {
        if (!isConnected) {
            toast.error('WebSocket is not connected. Please connect first.');
            return;
        }
        // Limpiar estados anteriores al enviar nueva solicitud
        setContacts(null);
        setChats(null);
        setOnlineUsers(null);
        setListError(null);

        // Send a generic 'list' type, specify actual type in payload
        const message = {
            type: 'list', // Use the generic type expected by the backend router
            payload: { 
                listType: listPayloadType // Specify the list type within the payload
            }
        };

        try {
            sendWebSocketMessage(message);
            toast.info(`Request sent: type='list', payload={listType: '${listPayloadType}'}.`);
        } catch (e) {
            console.error(`Error sending list request for '${listPayloadType}':`, e);
            toast.error(`Failed to send list request for '${listPayloadType}'. Check console.`);
            setListError(`Frontend error sending request: ${e instanceof Error ? e.message : String(e)}`);
        }
    };

    return (
        <div>
            <h2>List Requests (WS)</h2>
            <p>Envía solicitudes para obtener listas a través de WebSocket.</p>
            <p>Estado de WebSocket: <strong>{isConnected ? 'Conectado' : 'Desconectado'}</strong></p>
            {!token && <p style={{color: 'orange'}}>Advertencia: No hay token JWT. Necesitarás uno para conectarte.</p>}
            {!isConnected && token && <p style={{color: 'orange'}}>Advertencia: WebSocket desconectado. Ve a 'Connect (WS)'.</p>}

            <div style={{ marginTop: '1rem', marginBottom: '1rem' }}>
                <button
                    onClick={() => handleSendListRequest('contacts')}
                    disabled={!isConnected}
                    style={{ marginRight: '0.5rem' }}
                >
                    Get Contacts List
                </button>
                <button
                    onClick={() => handleSendListRequest('chats')}
                    disabled={!isConnected}
                     style={{ marginRight: '0.5rem' }}
               >
                    Get Chats List
                </button>
                <button
                    onClick={() => handleSendListRequest('online_users')}
                    disabled={!isConnected}
                >
                    Get Online Users List
                </button>
            </div>

            {/* Sección para mostrar resultados o errores */}            
            <div style={{ marginTop: '1rem', border: '1px solid #eee', padding: '1rem', minHeight: '200px', background: '#f8f8f8' }}>
                <h3>Response:</h3>
                {listError && (
                    <div style={{ color: 'red', marginBottom: '1rem' }}>
                        <strong>Error:</strong> {listError}
                    </div>
                )}
                {contacts && (
                    <div>
                        <h4>Contacts ({contacts.length})</h4>
                        <pre style={{ maxHeight: '300px', overflowY: 'auto', background: 'white', padding: '0.5rem' }}>
                            {JSON.stringify(contacts, null, 2)}
                        </pre>
                    </div>
                )}
                 {chats && (
                    <div>
                        <h4>Chats ({chats.length})</h4>
                        <pre style={{ maxHeight: '300px', overflowY: 'auto', background: 'white', padding: '0.5rem' }}>
                            {JSON.stringify(chats, null, 2)}
                        </pre>
                    </div>
                )}
                 {onlineUsers && (
                    <div>
                        <h4>Online Users ({onlineUsers.length})</h4>
                        <pre style={{ maxHeight: '300px', overflowY: 'auto', background: 'white', padding: '0.5rem' }}>
                            {JSON.stringify(onlineUsers, null, 2)}
                        </pre>
                    </div>
                )}
                {!listError && !contacts && !chats && !onlineUsers && (
                     <p>Click a button above to request a list.</p>
                )}
            </div>
        </div>
    );
};

export default TestListRequests; 