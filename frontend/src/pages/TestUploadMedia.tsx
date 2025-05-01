import React, { useState, useCallback } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { uploadMedia } from '../services/api';
import { toast } from 'react-toastify';

const TestUploadMedia: React.FC = () => {
    const { token } = useAuth();
    const [selectedFile, setSelectedFile] = useState<File | null>(null);
    const [loading, setLoading] = useState<boolean>(false);
    const [responseData, setResponseData] = useState<any>(null);
    const [previewUrl, setPreviewUrl] = useState<string | null>(null);

    const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setResponseData(null); // Limpiar respuesta anterior
        const file = event.target.files?.[0];
        if (file) {
            setSelectedFile(file);
            // Crear URL de previsualización
            const reader = new FileReader();
            reader.onloadend = () => {
                setPreviewUrl(reader.result as string);
            };
            reader.readAsDataURL(file);
        } else {
            setSelectedFile(null);
            setPreviewUrl(null);
        }
    };

    const handleSubmit = async (event: React.FormEvent) => {
        event.preventDefault();
        if (!token) {
            toast.warn('Please login first to upload media.');
            return;
        }
        if (!selectedFile) {
            toast.error('Please select a file to upload.');
            return;
        }

        setLoading(true);
        setResponseData(null);

        try {
            const data = await uploadMedia(selectedFile, token);
            setResponseData(data);
            toast.success('Media uploaded successfully!');
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'Unknown upload error';
            toast.error(`Upload Failed: ${errorMessage}`);
            setResponseData({ error: errorMessage });
        } finally {
            setLoading(false);
        }
    };

    // Limpiar selección
    const handleClear = useCallback(() => {
        setSelectedFile(null);
        setPreviewUrl(null);
        setResponseData(null);
        // Resetear el input file (forma un poco hacky)
        const fileInput = document.getElementById('mediaFile') as HTMLInputElement;
        if (fileInput) {
            fileInput.value = '';
        }
    }, []);

    return (
        <div>
            <h2>Upload Media (API)</h2>
            <p>Selecciona un archivo y súbelo al servidor (requiere token).</p>
            {!token && <p style={{ color: 'red' }}><strong>Necesitas iniciar sesión para usar esta función.</strong></p>}

            <form onSubmit={handleSubmit}>
                <div style={{ marginBottom: '1rem' }}>
                    <label htmlFor="mediaFile">Select File:</label>
                    <input
                        type="file"
                        id="mediaFile"
                        onChange={handleFileChange}
                        disabled={loading || !token}
                        style={{ display: 'block', marginTop: '0.5rem' }}
                        accept="image/*,video/*" // Aceptar imágenes y videos (ajustar si es necesario)
                    />
                </div>

                {previewUrl && selectedFile && (
                     <div style={{ marginBottom: '1rem' }}>
                        <p>Preview / Info:</p>
                        {selectedFile.type.startsWith('image/') ? (
                            <img src={previewUrl} alt="Preview" style={{ maxWidth: '200px', maxHeight: '200px', display: 'block' }} />
                        ) : (
                            <span>{selectedFile.name} ({selectedFile.type})</span>
                        )}
                    </div>
                )}

                <button type="submit" disabled={loading || !token || !selectedFile}>
                    {loading ? 'Uploading...' : 'Upload File'}
                </button>
                {selectedFile && (
                     <button type="button" onClick={handleClear} disabled={loading} style={{ marginLeft: '0.5rem', backgroundColor: '#888'}}>
                        Clear Selection
                    </button>
                )}
            </form>

            {responseData && (
                <div style={{ marginTop: '1rem' }}>
                    <h3>API Response:</h3>
                    <pre style={{ background: '#f4f4f4', padding: '1rem', borderRadius: '4px' }}>
                        {JSON.stringify(responseData, null, 2)}
                    </pre>
                </div>
            )}
        </div>
    );
};

export default TestUploadMedia; 