import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { loginUser } from '../services/api';
import { toast } from 'react-toastify';

const TestLogin: React.FC = () => {
    const [identifier, setIdentifier] = useState<string>(''); // Puede ser email o username
    const [password, setPassword] = useState<string>('');
    const [loading, setLoading] = useState<boolean>(false);
    const { setToken } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (event: React.FormEvent) => {
        event.preventDefault();
        setLoading(true);

        // Determinar si el identificador es un email
        const isEmail = identifier.includes('@');
        const credentials = {
            ...(isEmail ? { email: identifier } : { username: identifier }),
            password,
        };

        try {
            const data = await loginUser(credentials);
            setToken(data.token); // Guardar token en el contexto
            console.log('Login successful, token received:', data.token);
            toast.success('Login successful!');
            navigate('/ws/connect');
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'Unknown login error';
            toast.error(`Login Failed: ${errorMessage}`);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            <h2>Login (API)</h2>
            <p>Introduce tu email o nombre de usuario y contraseña para obtener un token JWT.</p>
            <form onSubmit={handleSubmit} style={{ maxWidth: '400px' }}>
                <div style={{ marginBottom: '1rem' }}>
                    <label htmlFor="identifier">Email o Username:</label>
                    <input
                        type="text"
                        id="identifier"
                        value={identifier}
                        onChange={(e) => setIdentifier(e.target.value)}
                        required
                        style={{ width: '100%', marginTop: '0.3rem' }}
                        disabled={loading}
                    />
                </div>
                <div style={{ marginBottom: '1rem' }}>
                    <label htmlFor="password">Password:</label>
                    <input
                        type="password"
                        id="password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        required
                        style={{ width: '100%', marginTop: '0.3rem' }}
                        disabled={loading}
                    />
                </div>
                <button type="submit" disabled={loading} style={{ width: '100%' }}>
                    {loading ? 'Logging in...' : 'Login'}
                </button>
            </form>
        </div>
    );
};

export default TestLogin; 