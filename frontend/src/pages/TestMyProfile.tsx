import React, { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { getMyProfile } from '../services/api'; // Asegúrate de que esta función esté exportada
import { toast } from 'react-toastify'; // Importar toast

const TestMyProfile: React.FC = () => {
    const { token } = useAuth();
    const [profileData, setProfileData] = useState<any>(null);
    const [loading, setLoading] = useState<boolean>(false);

    const handleFetchProfile = async () => {
        if (!token) {
            toast.warn('No token available. Please login first.'); // Usar toast.warn
            return;
        }
        setLoading(true);
        setProfileData(null);
        try {
            const data = await getMyProfile(token);
            setProfileData(data);
            toast.success('Profile data fetched successfully!'); // Toast de éxito
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'Unknown error fetching profile';
            toast.error(`Failed to fetch profile: ${errorMessage}`); // Mostrar error con toast
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            <h2>My Profile (API)</h2>
            {!token ? (
                <p>Necesitas hacer login (obtener un token) para ver tu perfil.</p>
            ) : (
                <div>
                     <p>Haz clic en el botón para obtener los datos de tu perfil usando el token actual.</p>
                     <button onClick={handleFetchProfile} disabled={loading}>
                        {loading ? 'Fetching...' : 'Get My Profile Data'}
                    </button>
                </div>
            )}

            {profileData && (
                <div style={{ marginTop: '1rem' }}>
                    <h3>Profile Data:</h3>
                    <pre style={{ background: '#f4f4f4', padding: '1rem', borderRadius: '4px' }}>
                        {JSON.stringify(profileData, null, 2)}
                    </pre>
                </div>
            )}
        </div>
    );
};

export default TestMyProfile; 