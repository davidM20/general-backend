import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import {
    sendWebSocketMessage,
    getWebSocketState,
    setWebSocketMessageHandler,
} from '../services/websocket';
import { toast } from 'react-toastify';

// Asumiendo que MyProfileResponse tiene la estructura del backend (adaptar si es necesario)
// Debería coincidir con la estructura MyProfileResponse en Go (types.go)
interface MyProfileResponse {
    id: number;
    firstName?: string;
    lastName?: string;
    userName: string;
    email: string;
    phone?: string;
    sex?: string;
    docId?: string;
    nationalityId?: number;
    nationalityName?: string;
    birthdate?: string; // La fecha puede venir como string
    picture?: string;
    degreeId?: number;
    degreeName?: string;
    universityId?: number;
    universityName?: string;
    roleId: number;
    roleName?: string;
    statusAuthorizedId: number;
    summary?: string;
    address?: string;
    github?: string;
    linkedin?: string;
    curriculum?: any; // Podríamos definir una interfaz más detallada para Curriculum
    isOnline?: boolean; // Este campo puede venir de handleGetProfile
}

// Tipo para el mensaje parseado
interface ParsedWebSocketMessage {
    type: string;
    payload: any;
    error?: string;
}

const TestMyProfileWS: React.FC = () => {
    const { token } = useAuth();
    const [isConnected, setIsConnected] = useState<boolean>(false);
    const [profileData, setProfileData] = useState<MyProfileResponse | null>(null);
    const [profileError, setProfileError] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState<boolean>(false);

    useEffect(() => {
        const checkConnection = () => {
            setIsConnected(getWebSocketState() === 1);
        };
        checkConnection();
        const intervalId = setInterval(checkConnection, 2000);

        const handleWsMessage = (event: MessageEvent) => {
            const parsedData = event.data as ParsedWebSocketMessage;
            console.log("Received Parsed WS Message (My Profile):", parsedData);

            switch (parsedData.type) {
                // Esperamos la respuesta específica del perfil
                case 'get_profile_response': // Este es el tipo que envía handleGetMyProfile/handleGetProfile
                    setIsLoading(false);
                    setProfileError(null);
                    setProfileData(parsedData.payload as MyProfileResponse);
                    toast.success("Profile data received!");
                    break;
                case 'error':
                    setIsLoading(false);
                    console.error('Error message from backend (My Profile):', parsedData.payload?.error || 'Unknown error');
                    setProfileError(parsedData.payload?.error || 'Received error from backend');
                    setProfileData(null);
                    toast.error("Error fetching profile.");
                    break;
                default:
                    // Ignorar otros mensajes
                    break;
            }
        };

        setWebSocketMessageHandler(handleWsMessage);

        return () => {
            clearInterval(intervalId);
            // setWebSocketMessageHandler(null); // Considerar desregistrar si causa problemas
        };
    }, []);

    const handleGetMyProfileRequest = () => {
        if (!isConnected) {
            toast.error('WebSocket is not connected.');
            return;
        }
        setIsLoading(true);
        setProfileData(null);
        setProfileError(null);

        const message = {
            type: 'get_my_profile', // El tipo de mensaje para solicitar el perfil propio
            payload: {} // No necesita payload
        };

        try {
            sendWebSocketMessage(message);
            toast.info('Request sent: get_my_profile.');
        } catch (e) {
            setIsLoading(false);
            console.error('Error sending get_my_profile request:', e);
            toast.error('Failed to send get_my_profile request. Check console.');
            setProfileError(`Frontend error sending request: ${e instanceof Error ? e.message : String(e)}`);
        }
    };

    return (
        <div>
            <h2>Get My Profile (WS)</h2>
            <p>Solicita y muestra tu perfil completo (incluyendo currículum) vía WebSocket.</p>
            <p>Estado de WebSocket: <strong>{isConnected ? 'Conectado' : 'Desconectado'}</strong></p>
            {!isConnected && <p style={{ color: 'orange' }}>Conéctate al WebSocket primero.</p>}

            <button
                onClick={handleGetMyProfileRequest}
                disabled={!isConnected || isLoading}
            >
                {isLoading ? 'Loading...' : 'Get My Profile'}
            </button>

            <div style={{ marginTop: '1rem', border: '1px solid #eee', padding: '1rem', minHeight: '300px', background: '#f8f8f8' }}>
                <h3>My Profile Data:</h3>
                {profileError && (
                    <div style={{ color: 'red', marginBottom: '1rem' }}>
                        <strong>Error:</strong> {profileError}
                    </div>
                )}
                {profileData && (
                    <pre style={{ maxHeight: '500px', overflowY: 'auto', background: 'white', padding: '0.5rem' }}>
                        {JSON.stringify(profileData, null, 2)}
                    </pre>
                )}
                {!profileError && !profileData && !isLoading && (
                    <p>Click the button above to request your profile.</p>
                )}
                 {isLoading && <p>Loading profile data...</p>}
            </div>
        </div>
    );
};

export default TestMyProfileWS; 