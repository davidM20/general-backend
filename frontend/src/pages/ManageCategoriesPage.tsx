import React, { useState, useEffect, useCallback } from 'react';
import { getCategories, addCategory } from '../services/api';
import { Category } from '../types/api';
import { toast } from 'react-toastify';
import axios from 'axios'; // Para type assertion en catch

const ManageCategoriesPage: React.FC = () => {
    const [categories, setCategories] = useState<Category[]>([]);
    const [newCategoryName, setNewCategoryName] = useState('');
    const [isLoadingList, setIsLoadingList] = useState(false);
    const [isAddingCategory, setIsAddingCategory] = useState(false);
    const [listError, setListError] = useState<string | null>(null);
    const [addError, setAddError] = useState<string | null>(null);

    const fetchCategories = useCallback(async () => {
        setIsLoadingList(true);
        setListError(null);
        try {
            const data = await getCategories();
            setCategories(data);
        } catch (error) {
            console.error("Failed to fetch categories:", error);
            setListError("Failed to load categories.");
            toast.error("Failed to load categories.");
        } finally {
            setIsLoadingList(false);
        }
    }, []);

    useEffect(() => {
        fetchCategories();
    }, [fetchCategories]);

    const handleAddCategory = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (!newCategoryName.trim()) {
            toast.warn("Category name cannot be empty.");
            return;
        }
        setIsAddingCategory(true);
        setAddError(null);
        try {
            const newCategory = await addCategory(newCategoryName.trim());
            setCategories(prev => [...prev, newCategory].sort((a, b) => a.name.localeCompare(b.name))); // Añadir y ordenar
            setNewCategoryName('');
            toast.success(`Category "${newCategory.name}" added successfully!`);
        } catch (error) {
            console.error("Failed to add category:", error);
            let errorMessage = "Failed to add category.";
             if (axios.isAxiosError(error) && error.response) {
                 if (error.response.status === 409) {
                     errorMessage = "Category with this name already exists.";
                 } else if (error.response.data?.error) {
                     errorMessage = error.response.data.error;
                 }
             }
            setAddError(errorMessage);
            toast.error(errorMessage);
        } finally {
            setIsAddingCategory(false);
        }
    };

    return (
        <div style={styles.container}>
            <h2>Manage Categories</h2>

            <div style={styles.section}>
                <h3>Existing Categories</h3>
                {isLoadingList && <p>Loading categories...</p>}
                {listError && <p style={styles.errorText}>{listError}</p>}
                {!isLoadingList && !listError && (
                    categories.length > 0 ? (
                        <ul style={styles.list}>
                            {categories.map(cat => (
                                <li key={cat.categoryId} style={styles.listItem}>
                                    {cat.name} (ID: {cat.categoryId})
                                    {/* Podríamos añadir botones de editar/eliminar aquí */}
                                </li>
                            ))}
                        </ul>
                    ) : (
                        <p>No categories found.</p>
                    )
                )}
            </div>

            <div style={styles.section}>
                <h3>Add New Category</h3>
                <form onSubmit={handleAddCategory} style={styles.form}>
                    <input
                        type="text"
                        value={newCategoryName}
                        onChange={(e) => setNewCategoryName(e.target.value)}
                        placeholder="New category name"
                        style={styles.input}
                        disabled={isAddingCategory}
                    />
                    <button type="submit" style={styles.button} disabled={isAddingCategory}>
                        {isAddingCategory ? 'Adding...' : 'Add Category'}
                    </button>
                </form>
                {addError && <p style={styles.errorText}>{addError}</p>}
            </div>
        </div>
    );
};

// Estilos básicos
const styles: { [key: string]: React.CSSProperties } = {
    container: { padding: '20px', fontFamily: 'Arial, sans-serif' },
    section: { marginBottom: '30px' },
    list: { listStyle: 'none', padding: 0 },
    listItem: { padding: '5px 0', borderBottom: '1px solid #eee' },
    form: { display: 'flex', gap: '10px', alignItems: 'center' },
    input: { padding: '8px 12px', border: '1px solid #ccc', borderRadius: '4px', flexGrow: 1 },
    button: {
        padding: '8px 15px', border: 'none', borderRadius: '4px',
        backgroundColor: '#28a745', color: 'white', cursor: 'pointer'
    },
    errorText: { color: 'red', marginTop: '10px' },
};

export default ManageCategoriesPage; 