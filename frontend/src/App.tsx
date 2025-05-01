import './App.css'
// import ApiTester from './components/ApiTester' // Ya no se usa directamente
// import WebSocketTester from './components/WebSocketTester' // Ya no se usa directamente
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Layout from './layouts/Layout';
import { AuthProvider } from './contexts/AuthContext';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

// Importar las páginas
import TestLogin from './pages/TestLogin';
import TestRegister from './pages/TestRegister';
import TestMyProfile from './pages/TestMyProfile';
import TestNationalities from './pages/TestNationalities';
import TestWebSocketConnect from './pages/TestWebSocketConnect';
import TestSendChatMessage from './pages/TestSendChatMessage';
import TestRegisterEnterprise from './pages/TestRegisterEnterprise';
import TestUploadMedia from './pages/TestUploadMedia';
import TestListRequests from './pages/TestListRequests';
import TestNotifications from './pages/TestNotifications';
// Placeholder para otras páginas
const PlaceholderPage: React.FC<{ title: string }> = ({ title }) => <div><h2>{title}</h2><p>Página en construcción...</p></div>;

function App() {
  // const [count, setCount] = useState(0)

  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes >
          <Route path="/" element={<Layout />}>
            {/* Ruta por defecto, redirige a login o a connect si hay token? */}
            <Route index element={<Navigate to="/login" replace />} />

            {/* Rutas API */}
            <Route path="login" element={<TestLogin />} />
            <Route path="register" element={<TestRegister />} />
            <Route path="profile" element={<TestMyProfile />} />
            <Route path="nationalities" element={<TestNationalities />} />
            <Route path="enterprise" element={<TestRegisterEnterprise />} />
            <Route path="upload" element={<TestUploadMedia />} />
            {/* Añadir más rutas API aquí */}

            {/* Rutas WebSocket */}
            <Route path="ws/connect" element={<TestWebSocketConnect />} />
            <Route path="ws/chat" element={<TestSendChatMessage />} />
            <Route path="ws/lists" element={<TestListRequests />} />
            <Route path="ws/notifications" element={<TestNotifications />} />
            {/* Añadir más rutas WS aquí */}

            {/* Ruta Catch-all para 404 dentro del layout */}
            <Route path="*" element={<div><h2>404 Not Found</h2><p>La página solicitada no existe.</p></div>} />
          </Route>
        </Routes>
      </BrowserRouter>
      <ToastContainer
        position="bottom-right"
        autoClose={5000}
        hideProgressBar={false}
        newestOnTop={false}
        closeOnClick
        rtl={false}
        pauseOnFocusLoss
        draggable
        pauseOnHover
        theme="colored"
        />
    </AuthProvider>
  )
}

export default App
