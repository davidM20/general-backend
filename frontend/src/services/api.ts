import axios from 'axios';
import { Category } from '../types/api'; // Importar el nuevo tipo

// Asume que el proxy inverso está corriendo en localhost:8080
const API_BASE_URL = 'http://localhost:8080/api/v1';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  // No incluimos el token aquí directamente, se añadirá por petición si es necesario
});

// Interceptor para añadir el token JWT a las solicitudes protegidas
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('authToken'); // O donde almacenes el token
  if (token && config.url !== '/login' && !config.url?.startsWith('/register')) { // No añadir token a login/registro
    // Asumiendo que el token se pasa vía URL para las rutas protegidas como hicimos con WS
    // Si la API espera el token en Header (más común):
    // config.headers.Authorization = `Bearer ${token}`;
    // Si es por URL para POST /categories:
     if (config.method === 'post' && config.url === '/categories') {
         config.url += `?token=${encodeURIComponent(token)}`;
     }
     // Si es por URL para otras rutas protegidas (ej GET /users/me):
     // else if (config.method === 'get' && config.url === '/users/me') { ... }
     // Es importante ser consistente con cómo el middleware espera el token
  }
  return config;
}, (error) => {
  return Promise.reject(error);
});

// Función de ejemplo para obtener nacionalidades (ruta pública)
export const getNationalities = async () => {
  try {
    const response = await apiClient.get('/nationalities');
    return response.data;
  } catch (error) {
    console.error('Error fetching nationalities:', error);
    throw error;
  }
};

// Función de ejemplo para obtener el perfil del usuario (ruta protegida)
// Necesita recibir el token JWT
export const getMyProfile = async (token: string) => {
    if (!token) {
        throw new Error('Token is required for protected routes');
    }
    try {
        // Pasamos el token como parámetro de URL
        const response = await apiClient.get('/users/me', {
            params: { token }
        });
        return response.data;
    } catch (error) {
        console.error('Error fetching profile:', error);
        throw error;
    }
};

// Función para hacer login
export const loginUser = async (credentials: { email?: string; username?: string; password?: string }) => {
    // Asegurarse de que al menos email o username estén presentes
    if ((!credentials.email && !credentials.username) || !credentials.password) {
        throw new Error('Email/Username and Password are required for login.');
    }
    try {
        const response = await apiClient.post('/login', credentials);
        // Asumimos que la respuesta exitosa tiene un campo 'token'
        if (response.data && response.data.token) {
            return response.data; // Devuelve { token: "..." }
        } else {
            throw new Error('Login successful but no token received.');
        }
    } catch (error: any) {
        console.error('Error during login:', error);
        // Intentar devolver el mensaje de error del backend si existe
        const errorMessage = error.response?.data?.error || error.message || 'Unknown login error';
        throw new Error(errorMessage);
    }
};

// Función para el primer paso del registro
export const registerUserStep1 = async (userData: {
    email: string;
    password?: string; // Hacer opcional si se genera en el backend o se pide después
    firstName: string;
    lastName: string;
    // Añadir otros campos si son necesarios en el paso 1
}) => {
    if (!userData.email || !userData.firstName || !userData.lastName) {
        throw new Error('Email, FirstName, and LastName are required for registration step 1.');
    }
    try {
        // Asumimos un endpoint /auth/register/step1
        // Puede que no requiera token
        const response = await apiClient.post('/auth/register/step1', userData);
        // La respuesta podría variar, ej: { success: true, userId: 123 } o solo un 200 OK.
        // Por ahora, solo devolvemos la data completa para inspección.
        return response.data;
    } catch (error: any) {
        console.error('Error during registration step 1:', error);
        const errorMessage = error.response?.data?.error || error.message || 'Unknown registration error';
        throw new Error(errorMessage);
    }
};

// Función para el segundo paso del registro
export const registerUserStep2 = async (userId: number, userData: {
    docId: string;
    nationalityId: number;
}) => {
    if (!userId || !userData.docId || !userData.nationalityId) {
        throw new Error('UserID, DocId, and NationalityId are required for registration step 2.');
    }
    try {
        // Endpoint: /auth/register/step2/{userId}
        const response = await apiClient.post(`/auth/register/step2/${userId}`, userData);
        return response.data; // Puede devolver un simple success o nada
    } catch (error: any) {
        console.error('Error during registration step 2:', error);
        const errorMessage = error.response?.data?.error || error.message || 'Unknown registration step 2 error';
        throw new Error(errorMessage);
    }
};

// Función para el tercer paso del registro
export const registerUserStep3 = async (userId: number, userData: {
    sex: string;
    birthdate: string; // Enviar como string YYYY-MM-DD
}) => {
    if (!userId || !userData.sex || !userData.birthdate) {
        throw new Error('UserID, Sex, and Birthdate are required for registration step 3.');
    }
    try {
        // Endpoint: /auth/register/step3/{userId}
        const response = await apiClient.post(`/auth/register/step3/${userId}`, userData);
        return response.data; // Puede devolver el usuario completo o un success
    } catch (error: any) {
        console.error('Error during registration step 3:', error);
        const errorMessage = error.response?.data?.error || error.message || 'Unknown registration step 3 error';
        throw new Error(errorMessage);
    }
};

// Función para registrar una empresa
export const registerEnterprise = async (enterpriseData: {
    rif: string;
    companyName: string;
    categoryId?: number; // Asumiendo que CategoryId es opcional o se puede obtener/seleccionar
    description?: string;
    location?: string;
    phone?: string;
}, token: string) => {
    if (!token) {
        throw new Error('Token is required to register an enterprise.');
    }
    if (!enterpriseData.rif || !enterpriseData.companyName) {
        throw new Error('RIF and Company Name are required.');
    }
    try {
        const response = await apiClient.post('/enterprises', enterpriseData, {
            params: { token } // Pasar token como parámetro URL
            // O si se pasa como Header:
            // headers: { Authorization: `Bearer ${token}` }
        });
        return response.data; // Asumimos que devuelve la empresa creada o un mensaje
    } catch (error: any) {
        console.error('Error registering enterprise:', error);
        const errorMessage = error.response?.data?.error || error.message || 'Unknown error registering enterprise';
        throw new Error(errorMessage);
    }
};

// Función para subir un archivo multimedia
export const uploadMedia = async (file: File, token: string) => {
    if (!token) {
        throw new Error('Token is required to upload media.');
    }
    if (!file) {
        throw new Error('File is required.');
    }

    const formData = new FormData();
    formData.append('media', file); // 'media' es el nombre esperado por el backend (ajustar si es diferente)

    try {
        const response = await apiClient.post('/media/upload', formData, {
             params: { token }, // Pasar token como parámetro URL
            headers: {
                'Content-Type': 'multipart/form-data', // Necesario para FormData
            },
            // O si se pasa como Header:
            // params: {}, // Sin token en params
            // headers: {
            //    'Content-Type': 'multipart/form-data',
            //    Authorization: `Bearer ${token}`
            // }
        });
        // Asume que la respuesta incluye MediaID y FileURL
        return response.data;
    } catch (error: any) {
        console.error('Error uploading media:', error);
        const errorMessage = error.response?.data?.error || error.message || 'Unknown error uploading media';
        throw new Error(errorMessage);
    }
};

// --- Funciones para Categorías ---

export const getCategories = async (): Promise<Category[]> => {
  try {
    const response = await apiClient.get<Category[]>('/categories');
    return response.data;
  } catch (error) {
    console.error('Error fetching categories:', error);
    throw error; // Re-lanzar para manejo en el componente
  }
};

export const addCategory = async (name: string): Promise<Category> => {
   // El token se añade automáticamente por el interceptor (vía URL en este caso)
   try {
        const response = await apiClient.post<Category>('/categories', { name }); // Ruta relativa a baseURL
        return response.data;
   } catch (error) {
       console.error('Error adding category:', error);
       throw error;
   }
};

// Aquí puedes añadir más funciones para otras rutas de la API
// Por ejemplo, para el registro:
/*
// ... registro paso 2, 3 ...
*/

export default apiClient; 