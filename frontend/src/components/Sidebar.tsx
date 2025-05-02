import React from 'react';
import { NavLink, Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import './Sidebar.css'; // Crearemos este archivo para estilos

const Sidebar: React.FC = () => {
    const { token, setToken } = useAuth();

    const handleLogout = () => {
        setToken(null);
        // Opcional: Desconectar WebSocket aquí si está conectado
        // disconnectWebSocket();
    };

    // Lista de endpoints API y acciones WS
    // TODO: Expandir esta lista con todas las rutas/acciones
    const apiEndpoints = [
        { path: '/login', name: 'Login (API)' },
        { path: '/register', name: 'Register (API)' }, // Podría ser una página con los 3 pasos
        { path: '/profile', name: 'My Profile (API)' },
        { path: '/nationalities', name: 'Nationalities (API)' },
        { path: '/enterprise', name: 'Register Enterprise (API)' },
        { path: '/upload', name: 'Upload Media (API)' },
        // ... añadir más rutas API
    ];

    const wsActions = [
        { path: '/ws/connect', name: 'Connect (WS)' }, // Una página para gestionar conexión/desconexión
        { path: '/ws/chat', name: 'Send Chat Message (WS)' },
        { path: '/ws/lists', name: 'List Requests (WS)' },
        { path: '/ws/notifications', name: 'Notifications (WS)' },
        { path: '/test-my-profile-ws', name: 'Get My Profile (WS)' },
        // ... añadir más acciones WS
    ];

    return (
        <div className="sidebar">
            <h3>API Endpoints</h3>
            <nav>
                <ul>
                    {apiEndpoints.map((endpoint) => (
                        <li key={endpoint.path}>
                            <NavLink
                                to={endpoint.path}
                                className={({ isActive }) => isActive ? "active-link" : ""}
                            >
                                {endpoint.name}
                            </NavLink>
                        </li>
                    ))}
                </ul>
            </nav>

            <h3>WebSocket Actions</h3>
            <nav>
                <ul>
                    {wsActions.map((action) => (
                        <li key={action.path}>
                            <NavLink
                                to={action.path}
                                className={({ isActive }) => isActive ? "active-link" : ""}
                            >
                                {action.name}
                            </NavLink>
                        </li>
                    ))}
                </ul>
            </nav>

            {token && (
                <div className="logout-section">
                    <button onClick={handleLogout}>Clear Token / Logout</button>
                </div>
            )}
        </div>
    );
};

export default Sidebar; 