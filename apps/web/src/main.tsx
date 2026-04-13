import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { Toaster } from "sonner";
import { Layout } from "./components/Layout";
import "./i18n";
import { DashboardPage } from "./pages/DashboardPage";
import { LoginPage } from "./pages/LoginPage";
import { SettingsUsersPage } from "./pages/SettingsUsersPage";
import { TracesPage } from "./pages/TracesPage";
import { TraceViewerPage } from "./pages/TraceViewerPage";
import { useAuthStore } from "./store/auth";
import "./styles.css";

const queryClient = new QueryClient();

function Private({ children }: { children: JSX.Element }) {
  const token = useAuthStore((s) => s.token);
  return token ? children : <Navigate to="/login" replace />;
}

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/"
            element={
              <Private>
                <Layout />
              </Private>
            }
          >
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="dashboard" element={<DashboardPage />} />
            <Route path="traces" element={<TracesPage />} />
            <Route path="traces/:traceId" element={<TraceViewerPage />} />
            <Route path="settings/users" element={<SettingsUsersPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
      <Toaster richColors position="top-right" closeButton />
    </QueryClientProvider>
  </React.StrictMode>,
);
