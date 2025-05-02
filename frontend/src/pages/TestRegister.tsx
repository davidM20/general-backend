import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { registerUserStep1, registerUserStep2, registerUserStep3, getNationalities } from '../services/api';
import { toast } from 'react-toastify';
import './TestRegister.css';

interface Nationality {
    id: number;
    countryName: string;
    // Add other fields if needed from the API response
}

const TestRegister: React.FC = () => {
    const [step, setStep] = useState<1 | 2 | 3>(1);
    const [formData, setFormData] = useState<any>({}); // Estado para guardar datos entre pasos
    const [userId, setUserId] = useState<number | null>(null); // Guardar ID del usuario después del paso 1
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [nationalities, setNationalities] = useState<Nationality[]>([]);
    const navigate = useNavigate();

    // Cargar nacionalidades para el paso 2
    useEffect(() => {
        const loadNationalities = async () => {
            try {
                const data = await getNationalities();
                setNationalities(data || []);
            } catch (error) {
                toast.error("Failed to load nationalities.");
                console.error("Error loading nationalities:", error);
            }
        };
        if (step === 2) {
            loadNationalities();
        }
    }, [step]);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
        const { name, value } = e.target;
        setFormData((prev: any) => ({ ...prev, [name]: value }));
    };

    const handleStep1Submit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsLoading(true);
        try {
            const step1Data = {
                firstName: formData.firstName,
                lastName: formData.lastName,
                userName: formData.userName, // Añadir si se pide en el backend
                email: formData.email,
                phone: formData.phone, // Añadir si se pide en el backend
                password: formData.password,
            };
            const response = await registerUserStep1(step1Data);
            toast.success('Step 1 completed successfully!');
            // Asumimos que la respuesta contiene el ID del usuario creado
            if (response && response.userId) { 
                setUserId(response.userId);
                setStep(2);
            } else {
                throw new Error('UserID not returned after step 1.');
            }
        } catch (error: any) {
            console.error("Step 1 Error:", error);
            toast.error(`Step 1 Failed: ${error.message}`);
        } finally {
            setIsLoading(false);
        }
    };

    const handleStep2Submit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!userId) {
            toast.error("UserID is missing. Cannot proceed.");
            return;
        }
        setIsLoading(true);
        try {
            const step2Data = {
                docId: formData.docId,
                nationalityId: parseInt(formData.nationalityId, 10), // Convertir a número
            };
            await registerUserStep2(userId, step2Data);
            toast.success('Step 2 completed successfully!');
            setStep(3);
        } catch (error: any) {
            console.error("Step 2 Error:", error);
            toast.error(`Step 2 Failed: ${error.message}`);
        } finally {
            setIsLoading(false);
        }
    };

    const handleStep3Submit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!userId) {
            toast.error("UserID is missing. Cannot proceed.");
            return;
        }
        setIsLoading(true);
        try {
            const step3Data = {
                sex: formData.sex,
                birthdate: formData.birthdate, // Asegurarse que el formato sea YYYY-MM-DD
            };
            await registerUserStep3(userId, step3Data);
            toast.success('Registration complete! Redirecting to login...');
            // Limpiar formulario y redirigir a login después de un breve delay
            setFormData({});
            setUserId(null);
            setStep(1);
            setTimeout(() => navigate('/login'), 2000);
        } catch (error: any) {
            console.error("Step 3 Error:", error);
            toast.error(`Step 3 Failed: ${error.message}`);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="register-form">
            <h2>Register (Step {step}/3)</h2>
            {step === 1 && (
                <form onSubmit={handleStep1Submit}>
                    <input type="text" name="firstName" placeholder="First Name" onChange={handleChange} required />
                    <input type="text" name="lastName" placeholder="Last Name" onChange={handleChange} required />
                    <input type="text" name="userName" placeholder="Username" onChange={handleChange} required />
                    <input type="email" name="email" placeholder="Email" onChange={handleChange} required />
                    <input type="text" name="phone" placeholder="Phone" onChange={handleChange} required />
                    <input type="password" name="password" placeholder="Password" onChange={handleChange} required />
                    <button type="submit" disabled={isLoading}>{isLoading ? 'Processing...' : 'Next Step'}</button>
                </form>
            )}
            {step === 2 && (
                <form onSubmit={handleStep2Submit}>
                    <input type="text" name="docId" placeholder="Document ID" onChange={handleChange} required />
                    <select name="nationalityId" onChange={handleChange} required defaultValue="">
                        <option value="" disabled>Select Nationality</option>
                        {nationalities.map(nat => (
                            <option key={nat.id} value={nat.id}>{nat.countryName}</option>
                        ))}
                    </select>
                    <button type="submit" disabled={isLoading}>{isLoading ? 'Processing...' : 'Next Step'}</button>
                    <button type="button" onClick={() => setStep(1)} disabled={isLoading}>Back</button>
                </form>
            )}
            {step === 3 && (
                <form onSubmit={handleStep3Submit}>
                    <select name="sex" onChange={handleChange} required defaultValue="">
                        <option value="" disabled>Select Sex</option>
                        <option value="Male">Male</option>
                        <option value="Female">Female</option>
                        <option value="Other">Other</option>
                    </select>
                    <input type="date" name="birthdate" placeholder="Birthdate" onChange={handleChange} required />
                    <button type="submit" disabled={isLoading}>{isLoading ? 'Registering...' : 'Complete Registration'}</button>
                    <button type="button" onClick={() => setStep(2)} disabled={isLoading}>Back</button>
                </form>
            )}
        </div>
    );
};

export default TestRegister; 