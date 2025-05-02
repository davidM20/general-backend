import React, { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../contexts/AuthContext';
import {
    addWebSocketListener,
    removeWebSocketListener,
    sendWebSocketMessage,
    getWebSocketState
} from '../services/websocket';
import {
    // Tipos para la solicitud
    // MessageTypeList,
    // ListRequestPayload,

    // Tipos de datos específicos y constantes de respuesta
    ContactInfo, ChatInfo, OnlineUserInfo,
    ListContactsResponsePayload, ListChatsResponsePayload, ListOnlineUsersResponsePayload,
    MessageTypeListContactsResponse,
    MessageTypeListChatsResponse,
    MessageTypeListOnlineUsersResponse,
    MessageTypeError // Para errores
} from '../types/websocket';
import { toast } from 'react-toastify';

// Tipo unión para los posibles datos de la lista
type ListDataType = ContactInfo[] | ChatInfo[] | OnlineUserInfo[];

const TestListRequests: React.FC = () => {
    const { token } = useAuth();
    const [listType, setListType] = useState<'contacts' | 'chats' | 'online_users'>('contacts');
    const [listData, setListData] = useState<ListDataType>([]); // Usar el tipo unión
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // Handler genérico para errores (se puede mover a un hook si se repite)
    const handleErrorResponse = useCallback((payload: { error: string } | any, errorMsg?: string) => {
        setIsLoading(false);
        const message = errorMsg || payload?.error || "An unknown WebSocket error occurred.";
        console.error("handleErrorResponse (List Requests) called:", message);
        setError(message);
        toast.error(message);
        setListData([]); // Limpiar datos en caso de error
    }, []);

    // Callbacks específicos para cada tipo de lista
    const handleContactsResponse = useCallback((payload: ListContactsResponsePayload, errorMsg?: string) => {
        console.log(`handleContactsResponse called:`, payload, "Error:", errorMsg);
        setIsLoading(false);
        if (errorMsg) {
            const message = `Failed to get contacts list: ${errorMsg}`;
            setError(message);
            toast.error(message);
            setListData([]);
        } else if (payload && payload.data) {
            setListData(payload.data); // Asignación directa es segura aquí
            setError(null);
            toast.success(`Contacts list received.`);
        } else {
             const message = `Received empty or invalid data for contacts list.`;
             setError(message);
             toast.warn(message);
             setListData([]);
        }
    }, []); // Dependencia vacía o setError, setListData, toast

    const handleChatsResponse = useCallback((payload: ListChatsResponsePayload, errorMsg?: string) => {
        console.log(`handleChatsResponse called:`, payload, "Error:", errorMsg);
        setIsLoading(false);
        if (errorMsg) {
            const message = `Failed to get chats list: ${errorMsg}`;
            setError(message);
            toast.error(message);
            setListData([]);
        } else if (payload && payload.data) {
            setListData(payload.data);
            setError(null);
            toast.success(`Chats list received.`);
        } else {
             const message = `Received empty or invalid data for chats list.`;
             setError(message);
             toast.warn(message);
             setListData([]);
        }
    }, []);

    const handleOnlineUsersResponse = useCallback((payload: ListOnlineUsersResponsePayload, errorMsg?: string) => {
        console.log(`handleOnlineUsersResponse called:`, payload, "Error:", errorMsg);
        setIsLoading(false);
        if (errorMsg) {
            const message = `Failed to get online users list: ${errorMsg}`;
            setError(message);
            toast.error(message);
            setListData([]);
        } else if (payload && payload.data) {
            setListData(payload.data);
            setError(null);
            toast.success(`Online users list received.`);
        } else {
             const message = `Received empty or invalid data for online users list.`;
             setError(message);
             toast.warn(message);
             setListData([]);
        }
    }, []);

    // Efecto para registrar/desregistrar listeners
    useEffect(() => {
        console.log("Registering WS listeners for list page...");
        addWebSocketListener(MessageTypeListContactsResponse, handleContactsResponse);
        addWebSocketListener(MessageTypeListChatsResponse, handleChatsResponse);
        addWebSocketListener(MessageTypeListOnlineUsersResponse, handleOnlineUsersResponse);
        addWebSocketListener(MessageTypeError, handleErrorResponse);

        // Función de limpieza al desmontar
        return () => {
            console.log("Unregistering WS listeners for list page...");
            removeWebSocketListener(MessageTypeListContactsResponse, handleContactsResponse);
            removeWebSocketListener(MessageTypeListChatsResponse, handleChatsResponse);
            removeWebSocketListener(MessageTypeListOnlineUsersResponse, handleOnlineUsersResponse);
            removeWebSocketListener(MessageTypeError, handleErrorResponse);
        };
        // Incluir todos los handlers en las dependencias
    }, [handleContactsResponse, handleChatsResponse, handleOnlineUsersResponse, handleErrorResponse]);

    // Función para solicitar la lista
    const fetchList = () => {
        const wsState = getWebSocketState();
        if (wsState !== WebSocket.OPEN) {
            toast.error("WebSocket is not connected. Cannot fetch list.");
            setError("WebSocket is not connected.");
            return;
        }
        setIsLoading(true);
        setError(null);
        setListData([]); // Limpiar datos anteriores
        console.log(`Sending 'list' request for ${listType} via WebSocket...`);
        sendWebSocketMessage({
            type: "list",
            payload: { listType: listType }
        });
        toast.info(`Request sent for ${listType} list.`);
    };

    return (
        <div>
            <h2>List Requests (WebSocket Test)</h2>
            <div style={{ marginBottom: '15px' }}>
                <label htmlFor="list-type-select">Select List Type: </label>
                <select
                    id="list-type-select"
                    value={listType}
                    onChange={(e) => {
                        setListType(e.target.value as 'contacts' | 'chats' | 'online_users');
                        setListData([]); // Limpiar datos al cambiar tipo
                        setError(null);
                    }}
                    disabled={isLoading}
                >
                    <option value="contacts">Contacts</option>
                    <option value="chats">Chats</option>
                    <option value="online_users">Online Users</option>
                </select>
                <button onClick={fetchList} disabled={isLoading || getWebSocketState() !== WebSocket.OPEN} style={{ marginLeft: '10px' }}>
                    {isLoading ? 'Loading...' : `Fetch ${listType.replace('_', ' ')}`}
                </button>
                 {getWebSocketState() !== WebSocket.OPEN && <p style={{color: 'orange', display: 'inline', marginLeft: '10px'}}>WebSocket Disconnected</p>}
            </div>

            {error && <p style={{ color: 'red' }}>Error: {error}</p>}

            <div style={{ marginTop: '20px', background: '#f9f9f9', padding: '15px', borderRadius: '5px', border: '1px solid #eee', minHeight: '200px' }}>
                <h3>List Data Received ({listType}):</h3>
                {listData.length > 0 ? (
                    <pre style={{ whiteSpace: 'pre-wrap', wordWrap: 'break-word' }}>
                        {JSON.stringify(listData, null, 2)}
                    </pre>
                ) : (
                     !isLoading && !error && getWebSocketState() === WebSocket.OPEN && <p>No data received or list is empty.</p>
                )}
                 {isLoading && <p>Loading list data...</p>}
            </div>
        </div>
    );
};

export default TestListRequests; 