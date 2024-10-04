'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import Dashboard from '@/components/dashboard';
import Login from '@/components/login';
import Link from 'next/link';
import React from 'react';

export default function Home() {
  const [isClient, setIsClient] = useState(false);
  const { isAuthenticated } = useAuth();

  useEffect(() => {
    setIsClient(true);
  }, []);

  if (!isClient) {
    return null; // or a loading spinner
  }

  return (
    <>
      {isAuthenticated ? (
        <Dashboard />
      ) : (
        <div>
          <Login />
          <p className="text-center mt-4">
            Don't have an account? <Link href="/register" className="text-blue-600 hover:underline">Sign up</Link>
          </p>
        </div>
      )}
    </>
  );
}