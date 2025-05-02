import React, { useState, useEffect, useCallback } from 'react';
import { addWebSocketListener, removeWebSocketListener, sendSearchRequest } from '../services/websocket';
import {
    MessageTypeSearchResponse,
    SearchResponsePayload, // Importamos el tipo que SÍ tiene 'results'
    TypedSearchResult, // Usaremos este tipo combinado
    TypedUserSearchResult,
    TypedEnterpriseSearchResult,
    MessageTypeError
} from '../types/websocket';
import { toast } from 'react-toastify';

const SearchPage: React.FC = () => {
    const [query, setQuery] = useState('');
    const [entityType, setEntityType] = useState<'user' | 'enterprise' | 'all'>('all');
    const [results, setResults] = useState<TypedSearchResult[]>([]); // Estado para resultados tipados
    const [isLoading, setIsLoading] = useState(false);
    const [searchAttempted, setSearchAttempted] = useState(false);

    const handleSearchResponse = useCallback((payload: SearchResponsePayload & { results: TypedSearchResult[] }, error?: string) => {
        setIsLoading(false);
        if (error) {
            toast.error(`Search failed: ${error}`);
            setResults([]);
        } else {
            console.log("Search results received:", payload.results);
            toast.success(`Found ${payload.results.length} results for "${payload.query}" (${payload.entityType}).`);
            setResults(payload.results);
        }
    }, []);

    const handleErrorResponse = useCallback((payload: any, error?: string) => {
        // Este handler podría ser más genérico, pero aquí lo usamos para errores de búsqueda
        // si no llegan por el SearchResponse
         if (error) { // El error podría venir en el campo 'error' genérico
             toast.error(`Error: ${error}`);
         } else if (payload?.error) { // O dentro del payload de un mensaje de tipo 'error'
             toast.error(`Error: ${payload.error}`);
         } else {
             toast.error("An unknown error occurred.");
         }
        setIsLoading(false);
        setResults([]);
    }, []);


    useEffect(() => {
        // Registrar listener al montar
        addWebSocketListener(MessageTypeSearchResponse, handleSearchResponse);
        addWebSocketListener(MessageTypeError, handleErrorResponse); // Escuchar errores genéricos también

        // Limpiar listener al desmontar
        return () => {
            removeWebSocketListener(MessageTypeSearchResponse, handleSearchResponse);
            removeWebSocketListener(MessageTypeError, handleErrorResponse);
        };
    }, [handleSearchResponse, handleErrorResponse]);

    const handleSearch = (e?: React.FormEvent<HTMLFormElement>) => {
        e?.preventDefault(); // Prevenir recarga de página si se usa en form
        if (!query.trim()) {
            toast.warn('Please enter a search term.');
            return;
        }
        setIsLoading(true);
        setSearchAttempted(true); // Marcar que se intentó buscar
        setResults([]); // Limpiar resultados anteriores
        console.log(`Sending search request: query="${query}", type="${entityType}"`);
        sendSearchRequest(query, entityType);
    };

    const renderResultItem = (result: TypedSearchResult) => {
        if (result.resultType === 'user') {
            const user = result as TypedUserSearchResult;
            return (
                <li key={`user-${user.id}`} style={styles.resultItem}>
                    <img src={user.picture || '/default-avatar.png'} alt={`${user.userName}'s picture`} style={styles.avatar} />
                    <div>
                        <strong>{user.firstName} {user.lastName}</strong> ({user.userName}) - {user.roleName || 'User'}
                        <br />
                        {user.universityName && <span>{user.degreeName} at {user.universityName}</span>}
                        {user.summary && <p style={{ fontSize: '0.9em', color: '#555' }}>{user.summary}</p>}
                    </div>
                    {/* Podríamos añadir un botón de 'Ver perfil' o 'Añadir contacto' */}
                </li>
            );
        } else if (result.resultType === 'enterprise') {
            const enterprise = result as TypedEnterpriseSearchResult;
            return (
                <li key={`enterprise-${enterprise.id}`} style={styles.resultItem}>
                     {/* Podríamos añadir un logo por defecto */}
                    <div>
                        <strong>{enterprise.companyName}</strong> (RIF: {enterprise.rif})
                        <br/>
                        {enterprise.categoryName && <span>Category: {enterprise.categoryName}</span>}
                        {enterprise.location && <span> | Location: {enterprise.location}</span>}
                        {enterprise.description && <p style={{ fontSize: '0.9em', color: '#555' }}>{enterprise.description}</p>}
                    </div>
                    {/* Podríamos añadir botón 'Ver detalles' */}
                </li>
            );
        }
        return null; // En caso de tipo desconocido
    };


    return (
        <div style={styles.container}>
            <h2>Search Users and Enterprises</h2>
            <form onSubmit={handleSearch} style={styles.form}>
                <input
                    type="text"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    placeholder="Search term..."
                    style={styles.input}
                    disabled={isLoading}
                />
                <select
                    value={entityType}
                    onChange={(e) => setEntityType(e.target.value as 'user' | 'enterprise' | 'all')}
                    style={styles.select}
                    disabled={isLoading}
                >
                    <option value="all">All</option>
                    <option value="user">Users</option>
                    <option value="enterprise">Enterprises</option>
                </select>
                <button type="submit" style={styles.button} disabled={isLoading}>
                    {isLoading ? 'Searching...' : 'Search'}
                </button>
            </form>

            <h3>Results:</h3>
             {isLoading && <p>Loading results...</p>}
             {!isLoading && searchAttempted && results.length === 0 && <p>No results found for "{query}".</p>}
             {!isLoading && results.length > 0 && (
                 <ul style={styles.resultsList}>
                     {results.map(renderResultItem)}
                 </ul>
             )}
        </div>
    );
};

// Estilos básicos para la página
const styles: { [key: string]: React.CSSProperties } = {
    container: {
        padding: '20px',
        fontFamily: 'Arial, sans-serif',
    },
    form: {
        display: 'flex',
        gap: '10px',
        marginBottom: '20px',
        alignItems: 'center',
    },
    input: {
        padding: '8px 12px',
        border: '1px solid #ccc',
        borderRadius: '4px',
        flexGrow: 1,
    },
    select: {
        padding: '8px 12px',
        border: '1px solid #ccc',
        borderRadius: '4px',
    },
    button: {
        padding: '8px 15px',
        border: 'none',
        borderRadius: '4px',
        backgroundColor: '#007bff',
        color: 'white',
        cursor: 'pointer',
    },
     resultsList: {
        listStyle: 'none',
        padding: 0,
        marginTop: '10px',
    },
    resultItem: {
        border: '1px solid #eee',
        borderRadius: '4px',
        padding: '15px',
        marginBottom: '10px',
        display: 'flex',
        alignItems: 'center',
        gap: '15px',
        backgroundColor: '#f9f9f9',
    },
    avatar: {
        width: '50px',
        height: '50px',
        borderRadius: '50%',
        objectFit: 'cover',
        border: '1px solid #ddd',
    }
};

export default SearchPage; 