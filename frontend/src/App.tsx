import './App.css'
// import ApiTester from './components/ApiTester' // Ya no se usa directamente
// import WebSocketTester from './components/WebSocketTester' // Ya no se usa directamente
import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import Layout from './layouts/Layout';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import SearchPage from './pages/SearchPage';
// import { WebSocketConnect } from './pages/WebSocketConnect'; // Comentado - Causa error

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
import TestMyProfileWS from './pages/TestMyProfileWS';
// Placeholder para otras páginas
const PlaceholderPage: React.FC<{ title: string }> = ({ title }) => <div><h2>{title}</h2><p>Página en construcción...</p></div>;

// Componente ProtectedRoute
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const { token } = useAuth();
    if (!token) {
        // Redirigir a login si no hay token
        return <Navigate to="/login" replace />;
    }
    return <>{children}</>; // Renderizar hijos si hay token
};

function App() {
  // const [count, setCount] = useState(0)

  return (
    <AuthProvider>
      <BrowserRouter>
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
        <Routes>
          {/* Rutas Públicas (sin Layout) */}
          <Route path="/login" element={<TestLogin />} />
          <Route path="/register" element={<TestRegister />} />

          {/* Rutas dentro del Layout Principal */}
          <Route path="/" element={<Layout />}>
            {/* Ruta índice redirige a una página por defecto (ej: /profile o /search si está logueado) */}
            {/* O simplemente mostrar un placeholder si no hay token */}
            <Route index element={<Navigate to="/login" replace />} /> {/* O redirigir a /profile si se prefiere como home logueado */}

            {/* Rutas de Prueba (accesibles desde Sidebar, algunas pueden necesitar protección) */}
            <Route path="profile" element={<TestMyProfile />} />
            <Route path="nationalities" element={<TestNationalities />} />
            <Route path="enterprise" element={<TestRegisterEnterprise />} />
            <Route path="upload" element={<TestUploadMedia />} />
            <Route path="ws/connect" element={<TestWebSocketConnect />} />
            <Route path="ws/chat" element={<TestSendChatMessage />} />
            <Route path="ws/lists" element={<TestListRequests />} />
            <Route path="ws/notifications" element={<TestNotifications />} />
            <Route path="test-my-profile-ws" element={<TestMyProfileWS />} />
            {/* Ruta ws-myprofile que apunta al mismo test que test-my-profile-ws */}
            <Route path="ws-myprofile" element={<TestMyProfileWS />} />

            {/* Ruta de Búsqueda Protegida */}
            {/* Se envuelve el elemento en ProtectedRoute */}
            <Route 
              path="search" 
              element={ 
                <ProtectedRoute>
                  <SearchPage />
                </ProtectedRoute>
              } 
            />

            {/* Placeholder para otras rutas mencionadas pero sin componente */}
            <Route path="home" element={<PlaceholderPage title="Home" />} />
            {/* <Route path="/profile" element={<ProfilePage />} /> */}
            {/* <Route path="/nationalities" element={<NationalitiesPage />} /> */}
            {/* <Route path="/register-enterprise" element={<RegisterEnterprisePage />} /> */}
            {/* <Route path="/upload-media" element={<UploadMediaPage />} /> */}
            {/* <Route path="/ws-chat" element={<WebSocketChat />} /> */}
            {/* <Route path="/ws-lists" element={<WebSocketLists />} /> */}
            {/* <Route path="/ws-notifications" element={<WebSocketNotifications />} /> */}

            {/* Catch-all 404 dentro del Layout */}
            <Route path="*" element={<PlaceholderPage title="404 Not Found" />} />
          </Route>

          {/* Podría haber un 404 fuera del Layout si se quiere */}
          {/* <Route path="*" element={<div><h2>404 Not Found</h2><p>Ruta global no encontrada.</p></div>} /> */}

        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}

export default App
