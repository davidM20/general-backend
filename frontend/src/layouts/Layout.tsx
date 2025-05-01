import React from 'react';
import { Outlet } from 'react-router-dom';
import Sidebar from '../components/Sidebar';
import './Layout.css'; // Estilos para el layout

const Layout: React.FC = () => {
    return (
        <div className="layout-container" style={{ display: 'flex', flex: 1 }}>
            <Sidebar />
            <main className="content-area">
                <Outlet /> {/* Aquí se renderizará el componente de la ruta activa */}
            </main>
        </div>
    );
};

export default Layout; 