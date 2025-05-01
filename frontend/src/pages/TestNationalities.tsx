import React, { useState, useEffect } from 'react';
import { getNationalities } from '../services/api';
// import { useAuth } from '../contexts/AuthContext'; // No necesario para esta API pública

const TestNationalities: React.FC = () => {
    // const { token } = useAuth(); // No se usa token aquí
    const [apiResponse, setApiResponse] = useState<any>(null);
    const [loading, setLoading] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);

    const handleGetNationalities = async () => {
        setLoading(true);
        setError(null);
        setApiResponse(null);
        try {
            const data = await getNationalities();
            setApiResponse(data);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Unknown API error');
        } finally {
            setLoading(false);
        }
    };

    // Llamar a la API al cargar la página
    useEffect(() => {
        handleGetNationalities();
    }, []);

    return (
        <div>
            <h2>Nationalities (API)</h2>
            <p>Esta página prueba la ruta pública <code>/misc/nationalities</code>.</p>
            <button onClick={handleGetNationalities} disabled={loading} style={{marginBottom: '1rem'}}>
                {loading ? 'Loading...' : 'Reload Nationalities'}
            </button>

            {/* Eliminamos la parte de probar el perfil */}
            {/* <hr style={{ margin: '1rem 0' }} /> ... */}

            {error && <div style={{ color: 'red', marginTop: '1rem' }}>Error: {error}</div>}

            {apiResponse && (
                <div style={{ marginTop: '1rem' }}>
                    <h3>API Response:</h3>
                    <pre style={{ background: '#f4f4f4', padding: '1rem', borderRadius: '4px' }}>
                        {JSON.stringify(apiResponse, null, 2)}
                    </pre>
                </div>
            )}
        </div>
    );
};

export default TestNationalities; 