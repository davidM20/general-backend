import React, { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { registerEnterprise } from '../services/api';
import { toast } from 'react-toastify';

interface EnterpriseData {
    rif: string;
    companyName: string;
    categoryId?: string; // Usar string para el input, convertir a number si es necesario
    description?: string;
    location?: string;
    phone?: string;
}

const TestRegisterEnterprise: React.FC = () => {
    const { token } = useAuth();
    const [formData, setFormData] = useState<EnterpriseData>({
        rif: '',
        companyName: '',
        categoryId: '',
        description: '',
        location: '',
        phone: '',
    });
    const [loading, setLoading] = useState<boolean>(false);
    const [responseData, setResponseData] = useState<any>(null);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    const handleSubmit = async (event: React.FormEvent) => {
        event.preventDefault();
        if (!token) {
            toast.warn('Please login first to register an enterprise.');
            return;
        }
        setLoading(true);
        setResponseData(null);

        // Preparar datos, convertir categoryId a número si existe
        const payload = {
            ...formData,
            categoryId: formData.categoryId ? parseInt(formData.categoryId, 10) : undefined,
        };
        // Validar si la conversión fue exitosa si se ingresó categoryId
        if (formData.categoryId && isNaN(payload.categoryId as number)) {
             toast.error('Invalid Category ID. Please enter a number or leave blank.');
             setLoading(false);
             return;
        }

        try {
            const data = await registerEnterprise(payload, token);
            setResponseData(data);
            toast.success('Enterprise registered successfully!');
            // Limpiar formulario?
            // setFormData({ rif: '', companyName: '', ... });
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'Unknown registration error';
            toast.error(`Registration Failed: ${errorMessage}`);
            setResponseData({ error: errorMessage });
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            <h2>Register Enterprise (API)</h2>
            <p>Registra una nueva empresa (requiere token de usuario).</p>
            {!token && <p style={{ color: 'red' }}><strong>Necesitas iniciar sesión para usar esta función.</strong></p>}

            <form onSubmit={handleSubmit} style={{ maxWidth: '500px' }}>
                <div style={{ marginBottom: '1rem' }}>
                    <label htmlFor="rif">RIF (Required):</label>
                    <input type="text" id="rif" name="rif" value={formData.rif} onChange={handleChange} required disabled={loading || !token} style={{ width: '100%', marginTop: '0.3rem' }}/>
                </div>
                <div style={{ marginBottom: '1rem' }}>
                    <label htmlFor="companyName">Company Name (Required):</label>
                    <input type="text" id="companyName" name="companyName" value={formData.companyName} onChange={handleChange} required disabled={loading || !token} style={{ width: '100%', marginTop: '0.3rem' }}/>
                </div>
                 <div style={{ marginBottom: '1rem' }}>
                    <label htmlFor="categoryId">Category ID (Optional, number):</label>
                    <input type="text" id="categoryId" name="categoryId" value={formData.categoryId} onChange={handleChange} disabled={loading || !token} style={{ width: '100%', marginTop: '0.3rem' }}/>
                </div>
                <div style={{ marginBottom: '1rem' }}>
                    <label htmlFor="description">Description:</label>
                    <textarea id="description" name="description" value={formData.description} onChange={handleChange} rows={3} disabled={loading || !token} style={{ width: '100%', marginTop: '0.3rem' }}/>
                </div>
                <div style={{ marginBottom: '1rem' }}>
                    <label htmlFor="location">Location:</label>
                    <input type="text" id="location" name="location" value={formData.location} onChange={handleChange} disabled={loading || !token} style={{ width: '100%', marginTop: '0.3rem' }}/>
                </div>
                 <div style={{ marginBottom: '1rem' }}>
                    <label htmlFor="phone">Phone:</label>
                    <input type="text" id="phone" name="phone" value={formData.phone} onChange={handleChange} disabled={loading || !token} style={{ width: '100%', marginTop: '0.3rem' }}/>
                </div>

                <button type="submit" disabled={loading || !token} style={{ width: '100%' }}>
                    {loading ? 'Registering...' : 'Register Enterprise'}
                </button>
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

export default TestRegisterEnterprise; 