import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Plus, Server, Globe, Power, Loader2, LogOut, RefreshCcw, Trash2, Edit2, X } from 'lucide-react';
import api from '../lib/api';
import { Button } from '../components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '../components/ui/card';

interface Project {
  ID: string;
  Name: string;
  Subdomain: string;
  ImageName: string;
  ContainerPort: number;
  Status: string;
  ContainerID: string;
}

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const [projects, setProjects] = useState<Project[]>([]);
  const [user, setUser] = useState<{username: string; base_domain?: string} | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<string | null>(null); // For start/stop/deploy spinner
  const [deleting, setDeleting] = useState<string | null>(null);
  
  // Modals
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  
  // Form State
  const [formData, setFormData] = useState({
    id: '',
    name: '',
    image: 'nginx:alpine',
    subdomain: '',
    port: 80
  });

  const fetchData = async () => {
      try {
           const [resProjects, resUser] = await Promise.all([
               api.get('/projects'),
               api.get('/auth/me').catch(() => ({ data: { username: 'Guest' } }))
           ]);
           setProjects(resProjects.data);
           setUser(resUser.data);
      } catch (err) {
          console.error("Failed to fetch data", err);
      } finally {
          setLoading(false);
      }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleLogout = () => {
      localStorage.removeItem('token');
      navigate('/login');
  };

  const handleAction = async (id: string, action: 'deploy' | 'start' | 'stop') => {
      setActionLoading(id);
      try {
          await api.post(`/projects/${id}/${action}`);
          // Reload projects to update status
          const res = await api.get('/projects');
          setProjects(res.data);
      } catch (err) {
          console.error(`${action} failed`, err);
          alert(`${action} failed`);
      } finally {
          setActionLoading(null);
      }
  };

  const handleDelete = async (id: string) => {
      if (!confirm("Are you sure you want to delete this project? This will stop and remove the container.")) return;
      setDeleting(id);
      try {
          await api.delete(`/projects/${id}`);
          const res = await api.get('/projects');
          setProjects(res.data);
      } catch (err) {
          console.error("Delete failed", err);
          alert("Failed to delete project");
      } finally {
          setDeleting(null);
      }
  };

  const openCreateModal = () => {
      setFormData({ id: '', name: '', image: 'nginx:alpine', subdomain: '', port: 80 });
      setShowCreateModal(true);
  };

  const openEditModal = (p: Project) => {
      setFormData({
          id: p.ID,
          name: p.Name,
          image: p.ImageName,
          subdomain: p.Subdomain,
          port: p.ContainerPort
      });
      setShowEditModal(true);
  };

  const handleCreateSubmit = async (e: React.FormEvent) => {
      e.preventDefault();
      try {
          await api.post('/projects', {
              ...formData,
              port: Number(formData.port)
          });
          setShowCreateModal(false);
          const res = await api.get('/projects');
          setProjects(res.data);
      } catch (err) {
          console.error("Create failed", err);
          alert("Failed to create project");
      }
  };

  const handleEditSubmit = async (e: React.FormEvent) => {
      e.preventDefault();
      try {
          await api.put(`/projects/${formData.id}`, {
              ...formData,
              port: Number(formData.port) // Ensure int
          });
          setShowEditModal(false);
          const res = await api.get('/projects');
          setProjects(res.data);
      } catch (err) {
          console.error("Update failed", err);
          alert("Failed to update project");
      }
  };

  return (
    <div className="min-h-screen bg-background text-foreground">
      {/* Navbar */}
      <nav className="border-b border-border bg-card/50 backdrop-blur-md sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
            <div className="flex items-center space-x-2">
                <Server className="h-6 w-6 text-primary" />
                <span className="font-bold text-xl tracking-tight">MultiTenant</span>
            </div>
            <div className="flex items-center space-x-4">
                <span className="text-sm text-muted-foreground mr-2">
                    Hi, {user?.username}
                </span>
                <Button variant="ghost" size="sm" onClick={handleLogout}>
                    <LogOut className="h-4 w-4 mr-2" />
                    Logout
                </Button>
            </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex items-center justify-between mb-8">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
                <p className="text-muted-foreground mt-1">Manage your deployed applications.</p>
            </div>
            <Button onClick={openCreateModal}>
                <Plus className="h-4 w-4 mr-2" />
                New Project
            </Button>
        </div>

        {loading ? (
             <div className="flex justify-center py-20">
                 <Loader2 className="h-8 w-8 animate-spin text-primary" />
             </div>
        ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {projects.map((project) => (
                    <motion.div
                        key={project.ID}
                        initial={{ opacity: 0, scale: 0.95 }}
                        animate={{ opacity: 1, scale: 1 }}
                        transition={{ duration: 0.2 }}
                    >
                        <Card className="hover:border-primary/50 transition-colors relative group">
                            {/* Action Buttons (Visible on Hover) */}
                            <div className="absolute top-3 right-3 flex space-x-1 opacity-100 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity bg-card/80 p-1 rounded-md">
                                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => openEditModal(project)}>
                                    <Edit2 className="h-3.5 w-3.5" />
                                </Button>
                                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive hover:text-destructive" onClick={() => handleDelete(project.ID)}>
                                    {deleting === project.ID ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Trash2 className="h-3.5 w-3.5" />}
                                </Button>
                            </div>

                            <CardHeader className="pb-3">
                                <div className="flex justify-between items-start">
                                    <CardTitle className="truncate pr-8">{project.Name}</CardTitle>
                                </div>
                                <div className="flex items-center space-x-2 mt-1">
                                     <span className={`px-2 py-0.5 rounded-full text-[10px] font-semibold uppercase tracking-wider ${
                                        project.Status === 'running' 
                                        ? 'bg-green-500/10 text-green-500 border border-green-500/20' 
                                        : 'bg-yellow-500/10 text-yellow-500 border border-yellow-500/20'
                                    }`}>
                                        {project.Status}
                                    </span>
                                </div>
                                <CardDescription className="text-xs font-mono bg-muted/50 p-1 rounded mt-2 truncate">
                                    Image: {project.ImageName}:{project.ContainerPort}
                                </CardDescription>
                            </CardHeader>
                            <CardContent className="pb-3">
                                <div className="space-y-2 text-sm">
                                    <div className="flex items-center text-muted-foreground p-2 rounded-md bg-secondary/50">
                                        <Globe className="h-4 w-4 mr-2 text-primary" />
                                        <a 
                                            href={`http://${project.Subdomain}.${user?.base_domain || 'localhost'}`} 
                                            target="_blank" 
                                            rel="noreferrer"
                                            className="hover:underline hover:text-primary transition-colors truncate"
                                        >
                                            {project.Subdomain}.{user?.base_domain || 'localhost'}
                                        </a>
                                    </div>
                                </div>
                            </CardContent>
                            <CardFooter className="pt-3 border-t border-border flex gap-2">
                                {project.Status === 'running' ? (
                                    <>
                                        <Button 
                                            variant="secondary" 
                                            size="sm" 
                                            className="flex-1 text-destructive hover:text-destructive"
                                            onClick={() => handleAction(project.ID, 'stop')}
                                            disabled={actionLoading === project.ID}
                                        >
                                            {actionLoading === project.ID ? <Loader2 className="h-4 w-4 animate-spin" /> : <Power className="h-4 w-4 mr-2" />}
                                            Stop
                                        </Button>
                                        <Button 
                                            size="sm" 
                                            variant="secondary" 
                                            onClick={() => handleAction(project.ID, 'deploy')} 
                                            disabled={actionLoading === project.ID}
                                        >
                                            {actionLoading === project.ID ? <Loader2 className="h-4 w-4 animate-spin"/> : <RefreshCcw className="h-4 w-4" />}
                                        </Button>
                                    </>
                                ) : (
                                    <>
                                        <Button 
                                            variant="secondary" 
                                            size="sm" 
                                            className="flex-1 text-green-500 hover:text-green-600"
                                            onClick={() => handleAction(project.ID, 'start')}
                                            disabled={actionLoading === project.ID}
                                        >
                                            {actionLoading === project.ID ? <Loader2 className="h-4 w-4 animate-spin" /> : <Power className="h-4 w-4 mr-2" />}
                                            Start
                                        </Button>
                                         <Button 
                                            size="sm" 
                                            variant="default" // Primary color for deploy if not running
                                            onClick={() => handleAction(project.ID, 'deploy')} 
                                            disabled={actionLoading === project.ID}
                                        >
                                            {actionLoading === project.ID ? <Loader2 className="h-4 w-4 animate-spin"/> : <RefreshCcw className="h-4 w-4" />}
                                        </Button>
                                    </>
                                )}
                            </CardFooter>
                        </Card>
                    </motion.div>
                ))}

                {projects.length === 0 && (
                    <div className="col-span-full text-center py-20 border border-dashed border-border rounded-lg bg-card/30">
                        <Server className="h-12 w-12 mx-auto text-muted-foreground mb-4 opacity-50" />
                        <h3 className="text-lg font-medium">No projects found</h3>
                        <p className="text-muted-foreground mb-4">Get started by creating your first application.</p>
                        <Button onClick={openCreateModal}>Create Project</Button>
                    </div>
                )}
            </div>
        )}
      </main>

      {/* Create/Edit Modal Overlay */}
      {(showCreateModal || showEditModal) && (
          <div className="fixed inset-0 z-[100] flex items-center justify-center bg-background/80 backdrop-blur-sm p-4 animate-in fade-in duration-200">
              <motion.div 
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                className="w-full max-w-lg"
              >
                  <Card className="shadow-2xl border-primary/20">
                      <CardHeader className="relative">
                          <Button 
                            variant="ghost" 
                            size="icon" 
                            className="absolute right-4 top-4 h-8 w-8" 
                            onClick={() => { setShowCreateModal(false); setShowEditModal(false); }}
                          >
                            <X className="h-4 w-4" />
                          </Button>
                          <CardTitle>{showEditModal ? 'Edit Project' : 'New Project'}</CardTitle>
                          <CardDescription>
                              {showEditModal ? 'Update configuration (requires change redeploy).' : 'Deploy a new container instance.'}
                          </CardDescription>
                      </CardHeader>
                      <form onSubmit={showEditModal ? handleEditSubmit : handleCreateSubmit}>
                          <CardContent className="space-y-4">
                              <div className="space-y-2">
                                  <label className="text-sm font-medium">Project Name</label>
                                  <input 
                                    className="flex h-10 w-full rounded-md border border-input bg-background/50 px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
                                    placeholder="My Awesome App"
                                    value={formData.name}
                                    onChange={e => setFormData({...formData, name: e.target.value})}
                                    required
                                  />
                              </div>
                              <div className="space-y-2">
                                  <label className="text-sm font-medium">Subdomain (URL)</label>
                                  <div className="flex items-center">
                                      <input 
                                        className="flex h-10 w-full rounded-l-md border border-input bg-background/50 px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
                                        placeholder="myapp"
                                        value={formData.subdomain}
                                        onChange={e => setFormData({...formData, subdomain: e.target.value})}
                                        required
                                      />
                                      <div className="h-10 px-3 flex items-center bg-muted border border-l-0 border-input rounded-r-md text-sm text-muted-foreground font-mono">
                                          .{user?.base_domain || 'localhost'}
                                      </div>
                                  </div>
                              </div>
                              <div className="grid grid-cols-2 gap-4">
                                  <div className="space-y-2">
                                      <label className="text-sm font-medium">Docker Image</label>
                                      <input 
                                        className="flex h-10 w-full rounded-md border border-input bg-background/50 px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary font-mono"
                                        value={formData.image}
                                        onChange={e => setFormData({...formData, image: e.target.value})}
                                        required
                                        placeholder="e.g. nginx:alpine"
                                      />
                                  </div>
                                  <div className="space-y-2">
                                      <label className="text-sm font-medium">Container Port</label>
                                      <input 
                                        type="number"
                                        className="flex h-10 w-full rounded-md border border-input bg-background/50 px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary font-mono"
                                        value={formData.port}
                                        onChange={e => setFormData({...formData, port: Number(e.target.value)})}
                                        required
                                        placeholder="80"
                                      />
                                  </div>
                              </div>
                          </CardContent>
                          <CardFooter className="flex justify-end space-x-2 bg-muted/20 py-4">
                              <Button type="button" variant="ghost" onClick={() => { setShowCreateModal(false); setShowEditModal(false); }}>Cancel</Button>
                              <Button type="submit" disabled={loading}>
                                {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                {showEditModal ? 'Save Changes' : 'Create Project'}
                              </Button>
                          </CardFooter>
                      </form>
                  </Card>
              </motion.div>
          </div>
      )}
    </div>
  );
};

export default Dashboard;
