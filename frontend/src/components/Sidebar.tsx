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
        { path: '/ws-myprofile', name: 'My Profile (WS)' },
        { path: '/search', name: 'Search (WS)' },
        // ... añadir más acciones WS
    ];

    // TODO: Podríamos separar las rutas de prueba de las rutas "reales" protegidas
    const testRoutes = [
         // ... (lista actual de apiEndpoints y wsActions puede ir aquí o mantenerse separada)
         { path: '/ws/connect', name: 'Connect (WS Test)' },
         { path: '/ws/chat', name: 'Send Chat (WS Test)' },
         { path: '/ws/lists', name: 'List Requests (WS Test)' },
         { path: '/ws/notifications', name: 'Notifications (WS Test)' },
         { path: '/test-my-profile-ws', name: 'Get My Profile (WS Test)' },
         { path: '/profile', name: 'My Profile (API Test)' },
         { path: '/nationalities', name: 'Nationalities (API Test)' },
         { path: '/enterprise', name: 'Register Enterprise (API Test)' },
         { path: '/upload', name: 'Upload Media (API Test)' },
    ];

     const protectedRoutes = [ // Rutas que requieren login
         { path: '/search', name: 'Search (WS)' },
         { path: '/manage-categories', name: 'Manage Categories' }, // Nuevo enlace
         // Añadir más rutas protegidas aquí
     ];

    return (
        <div className="sidebar">
            <h3>Testing Area</h3>
            <nav>
                <ul>
                    {testRoutes.map((route) => (
                        <li key={route.path}>
                            <NavLink
                                to={route.path}
                                className={({ isActive }) => isActive ? "active-link" : ""}
                            >
                                {route.name}
                            </NavLink>
                        </li>
                    ))}
                </ul>
            </nav>

            <h3>App Features</h3>
            <nav>
                <ul>
                    {protectedRoutes.map((route) => (
                        <li key={route.path}>
                            <NavLink
                                to={route.path}
                                className={({ isActive }) => isActive ? "active-link" : ""}
                            >
                                {route.name}
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