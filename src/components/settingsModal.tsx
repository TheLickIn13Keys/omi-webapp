"use client"

import { useState } from 'react'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useToast } from "@/hooks/use-toast"

export default function SettingsModal() {
  const [gcpCredentialsFile, setGcpCredentialsFile] = useState<File | null>(null)
  const [gcpBucketName, setGcpBucketName] = useState('')
  const [gladiaKey, setGladiaKey] = useState('')
  const [voiceRecording, setVoiceRecording] = useState<File | null>(null)
  const { toast } = useToast()

  const handleGcpCredentialsFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (file) {
      setGcpCredentialsFile(file)
    }
  }

  const handleVoiceRecordingUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (file) {
      setVoiceRecording(file)
    }
  }

  const saveGCPCredentials = async () => {
    if (!gcpCredentialsFile) {
      toast({
        title: "Error",
        description: "Please upload a GCP credentials file",
        variant: "destructive",
      })
      return
    }

    try {
      const fileContent = await gcpCredentialsFile.text()
      const base64Credentials = btoa(fileContent)

      const response = await fetch('https://aggieworks-backend.server.bardia.app/gcp-credentials', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({
          credentials: base64Credentials,
          bucket_name: gcpBucketName,
          gladia_key: gladiaKey
        })
      });

      if (response.ok) {
        toast({
          title: "Success",
          description: "GCP credentials and Gladia key saved successfully",
        })
      } else {
        throw new Error('Failed to save GCP credentials and Gladia key');
      }
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to save GCP credentials and Gladia key",
        variant: "destructive",
      })
    }
  }

  return (
    <Tabs defaultValue="bucket-info">
      <TabsList className="grid w-full grid-cols-2">
        <TabsTrigger value="bucket-info">Bucket Info</TabsTrigger>
        <TabsTrigger value="voice-recording">Voice Recording</TabsTrigger>
      </TabsList>
      <TabsContent value="bucket-info" className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="gcp-credentials">GCP Credentials File</Label>
          <Input
            id="gcp-credentials"
            type="file"
            onChange={handleGcpCredentialsFileUpload}
            accept=".json"
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="gcp-bucket-name">GCP Bucket Name</Label>
          <Input
            id="gcp-bucket-name"
            value={gcpBucketName}
            onChange={(e) => setGcpBucketName(e.target.value)}
            placeholder="Enter your GCP bucket name"
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="gladia-key">Gladia API Key</Label>
          <Input
            id="gladia-key"
            value={gladiaKey}
            onChange={(e) => setGladiaKey(e.target.value)}
            placeholder="Enter your Gladia API key"
          />
        </div>
        //TODO: give feedback when clicked
        <Button className="w-full" onClick={saveGCPCredentials}>Save Bucket Info</Button>
      </TabsContent>
      <TabsContent value="voice-recording" className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="voice-recording">Upload Voice Recording (WIP)</Label>
          <Input
            id="voice-recording"
            type="file"
            accept="audio/*"
            onChange={handleVoiceRecordingUpload}
          />
        </div>
        <p className="text-sm text-muted-foreground">
          Upload a recording of your voice to improve person detection. This feature is currently a work in progress.
        </p>
        <Button className="w-full" disabled>Upload Voice Recording (WIP)</Button>
      </TabsContent>
    </Tabs>
  )
}