"use client"


import { useState } from 'react'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Slider } from "@/components/ui/slider"
import { Switch } from "@/components/ui/switch"
import { Music, Play, Send } from 'lucide-react'
import React from 'react'

interface Conversation {
  name: string;
  // Add other properties of the conversation object if needed
}

export default function ConversationView({ conversation }: { conversation: Conversation }) {
  const [chatMessage, setChatMessage] = useState('')

  return (
    <>
      <h2 className="text-2xl font-bold">{conversation.name}</h2>
      
      {/* Timeline */}
      <Card>
        <CardHeader>
          <CardTitle>Timeline</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <Slider defaultValue={[0]} max={100} step={1} />
            <div className="flex justify-between">
              <Button size="icon" variant="outline">
                <Play className="h-4 w-4" />
              </Button>
              <span>00:00 / 10:00</span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Transcript and Chat */}
      <Tabs defaultValue="transcript" className="space-y-4">
        <TabsList>
          <TabsTrigger value="transcript">Transcript</TabsTrigger>
          <TabsTrigger value="chat">Chat</TabsTrigger>
        </TabsList>
        <TabsContent value="transcript" className="space-y-4">
          <Card>
            <CardContent className="p-4">
              <ScrollArea className="h-[300px]">
                <div className="space-y-4">
                  <p><strong>Speaker 1:</strong> Hello, how are you today?</p>
                  <p><strong>Speaker 2:</strong> I'm doing well, thank you. How about you?</p>
                  <p><strong>Speaker 1:</strong> I'm great, thanks for asking!</p>
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="chat" className="space-y-4">
          <Card>
            <CardContent className="p-4">
              <ScrollArea className="h-[300px]">
                <div className="space-y-4">
                  <p><strong>You:</strong> What was the main topic of the conversation?</p>
                  <p><strong>AI:</strong> The main topic of the conversation was a friendly greeting and exchange of pleasantries.</p>
                </div>
              </ScrollArea>
            </CardContent>
          </Card>
          <div className="flex space-x-2">
            <Input
              type="text"
              placeholder="Ask about your recordings..."
              value={chatMessage}
              onChange={(e) => setChatMessage(e.target.value)}
            />
            <Button size="icon">
              <Send className="h-4 w-4" />
            </Button>
          </div>
        </TabsContent>
      </Tabs>

      {/* Plugin Outputs */}
      <Card>
        <CardHeader>
          <CardTitle>Plugin Outputs</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-4">
            <Card>
              <CardHeader>
                <CardTitle className="text-sm">Sentiment Analysis</CardTitle>
              </CardHeader>
              <CardContent>
                <p>Positive: 80%</p>
                <p>Neutral: 15%</p>
                <p>Negative: 5%</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle className="text-sm">Bias Detection</CardTitle>
              </CardHeader>
              <CardContent>
                <p>No significant bias detected</p>
              </CardContent>
            </Card>
          </div>
        </CardContent>
      </Card>

      {/* Music Analyzer */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Music Analyzer</CardTitle>
          <Switch />
        </CardHeader>
        <CardContent>
          <div className="flex items-center space-x-4">
            <Music className="h-4 w-4" />
            <div>
              <p className="text-sm font-medium">Detected Song</p>
              <p className="text-xs text-muted-foreground">Artist - Song Name</p>
            </div>
            <span className="text-xs text-muted-foreground">00:15 - 00:45</span>
          </div>
        </CardContent>
      </Card>
    </>
  )
}