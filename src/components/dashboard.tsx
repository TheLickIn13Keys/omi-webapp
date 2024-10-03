"use client"


import { useState } from 'react'

// Define the Conversation type
type Conversation = {
  id: string;
  name: string;
};
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card } from "@/components/ui/card"
import { Dialog, DialogContent, DialogTrigger } from "@/components/ui/dialog"
import Sidebar from './sidebar'
import ActionItemsSummary from './globalActionItems'
import ConversationView from './conversationView'
import SettingsModal from './settingsModal'
import PluginsMarketplace from './plugins'
import { Cog, Search } from 'lucide-react'
import React from 'react';

export default function Dashboard() {
  const [globalSearch, setGlobalSearch] = useState('')
  const [selectedConversation, setSelectedConversation] = useState<Conversation | null>(null)
  const [isAuthenticated, setIsAuthenticated] = useState(true)
  const [showPluginsMarketplace, setShowPluginsMarketplace] = useState(false)

  const conversations = [
    { id: "1", name: "Morning Coffee Chat" },
    { id: "2", name: "Team Meeting Notes" },
    { id: "3", name: "Client Presentation Prep" },
  ]

  const handleLogout = () => {
    setIsAuthenticated(false)
  }

  const togglePluginsMarketplace = () => {
    setShowPluginsMarketplace(!showPluginsMarketplace)
  }

  if (!isAuthenticated) {
    return null
  }

  return (
    <div className="flex h-screen bg-gray-100">
      <Sidebar
        conversations={conversations}
        onConversationSelect={setSelectedConversation}
        onLogout={handleLogout}
        onPluginsClick={togglePluginsMarketplace}
      />
      <div className="flex-1 p-6 space-y-6 overflow-auto">
        {showPluginsMarketplace ? (
          <PluginsMarketplace onClose={togglePluginsMarketplace} />
        ) : (
          <>
            <div className="flex space-x-2">
              <Input
                type="text"
                placeholder="Search all transcripts..."
                value={globalSearch}
                onChange={(e) => setGlobalSearch(e.target.value)}
                className="flex-grow"
              />
              <Button size="icon">
                <Search className="h-4 w-4" />
              </Button>
            </div>

            <ActionItemsSummary />

            {selectedConversation ? (
              <ConversationView conversation={selectedConversation} />
            ) : (
              <div className="flex items-center justify-center h-[calc(100vh-200px)]">
                <p className="text-xl text-gray-500">Click a conversation to get started</p>
              </div>
            )}
          </>
        )}
      </div>
      <Dialog>
        <DialogTrigger asChild>
          <Button variant="ghost" size="icon" className="absolute bottom-4 right-4">
            <Cog className="h-5 w-5" />
          </Button>
        </DialogTrigger>
        <DialogContent className="sm:max-w-[425px]">
          <SettingsModal />
        </DialogContent>
      </Dialog>
    </div>
  )
}