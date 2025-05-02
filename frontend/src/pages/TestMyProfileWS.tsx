import React, { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../contexts/AuthContext';
import {
    addWebSocketListener,
    removeWebSocketListener,
    sendWebSocketMessage,
    getWebSocketState,
} from '../services/websocket';
import {
    MessageTypeGetProfileResponse,
    MyProfileResponse,
    MessageTypeError
} from '../types/websocket';
import { toast } from 'react-toastify';

// Eliminar cualquier definición local conflictiva de MyProfileResponse si existe
// interface MyProfileResponse { ... } // <-- ¡Eliminar esto si existe!

// Tipo para el mensaje parseado
interface ParsedWebSocketMessage {
    type: string;
    payload: any;
    error?: string;
}

const TestMyProfileWS: React.FC = () => {
    const { token } = useAuth();
    const [profileData, setProfileData] = useState<MyProfileResponse | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // Callback para manejar la respuesta del perfil
    const handleProfileResponse = useCallback((payload: MyProfileResponse, errorMsg?: string) => {
        console.log("handleProfileResponse called with payload:", payload, "Error:", errorMsg);
        setIsLoading(false);
        if (errorMsg) {
            const message = `Failed to get profile: ${errorMsg}`;
            setError(message);
            toast.error(message);
            setProfileData(null);
        } else if (payload) {
            setProfileData(payload);
            setError(null);
            toast.success("Profile data received.");
            console.log("Profile Data State Updated (inside handleProfileResponse):", payload);
        } else {
             const message = "Received empty profile data.";
             setError(message);
             toast.warn(message);
             setProfileData(null);
        }
    }, []);

    // Callback para manejar errores generales de WS
    const handleErrorResponse = useCallback((payload: { error: string } | any, errorMsg?: string) => {
        setIsLoading(false);
        const message = errorMsg || payload?.error || "An unknown WebSocket error occurred.";
        console.error("handleErrorResponse called:", message);
        setError(message);
        toast.error(message);
        setProfileData(null);
    }, []);

    // Efecto para registrar/desregistrar listeners y pedir datos al montar
    useEffect(() => {
        console.log("Registering WS listeners for profile page...");
        addWebSocketListener(MessageTypeGetProfileResponse, handleProfileResponse);
        addWebSocketListener(MessageTypeError, handleErrorResponse);

        // Solicitar datos del perfil cuando el componente se monta si WS está conectado
        if (getWebSocketState() === WebSocket.OPEN) {
             fetchProfile();
        } else {
             setError("WebSocket not connected on mount.");
             // Podrías intentar conectar aquí o mostrar un mensaje para conectar manualmente
        }

        // Función de limpieza al desmontar
        return () => {
            console.log("Unregistering WS listeners for profile page...");
            removeWebSocketListener(MessageTypeGetProfileResponse, handleProfileResponse);
            removeWebSocketListener(MessageTypeError, handleErrorResponse);
        };
    }, [handleProfileResponse, handleErrorResponse]);

    // Función para solicitar el perfil
    const fetchProfile = () => {
        const wsState = getWebSocketState();
        if (wsState !== WebSocket.OPEN) {
            toast.error("WebSocket is not connected. Cannot fetch profile.");
            setError("WebSocket is not connected.");
            return;
        }
        setIsLoading(true);
        setError(null);
        setProfileData(null); // Limpiar datos anteriores
        console.log("Sending 'get_my_profile' request via WebSocket...");
        sendWebSocketMessage({ type: "get_my_profile" });
    };

    return (
        <div>
            <h2>My Profile (WebSocket Test)</h2>
            <button onClick={fetchProfile} disabled={isLoading || getWebSocketState() !== WebSocket.OPEN}>
                {isLoading ? 'Loading Profile...' : 'Fetch My Profile'}
            </button>
             {getWebSocketState() !== WebSocket.OPEN && <p style={{color: 'orange'}}>WebSocket Disconnected</p>}

            {error && <p style={{ color: 'red', marginTop: '10px' }}>Error: {error}</p>}

            {profileData ? (
                <div style={{ marginTop: '20px', background: '#f9f9f9', padding: '15px', borderRadius: '5px', border: '1px solid #eee' }}>
                    <h3>Profile Data Received:</h3>
                    <pre style={{ whiteSpace: 'pre-wrap', wordWrap: 'break-word' }}>
                        {JSON.stringify(profileData, null, 2)}
                    </pre>
                </div>
            ) : (
                !isLoading && !error && getWebSocketState() === WebSocket.OPEN && <p style={{ marginTop: '10px' }}>Click the button to fetch profile data.</p>
            )}
        </div>
    );
};

export default TestMyProfileWS; 