"use client"

import { useState } from 'react'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

export default function SettingsModal() {
  const [gcpCredentials, setGcpCredentials] = useState('')
  const [gcpBucketName, setGcpBucketName] = useState('')
  const [voiceRecording, setVoiceRecording] = useState<File | null>(null)

  const handleVoiceRecordingUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (file) {
      setVoiceRecording(file)
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
          <Label htmlFor="gcp-credentials">GCP Credentials (Base64)</Label>
          <Input
            id="gcp-credentials"
            value={gcpCredentials}
            onChange={(e) => setGcpCredentials(e.target.value)}
            placeholder="Enter your GCP credentials"
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
        <Button className="w-full">Save Bucket Info</Button>
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