# 🎙️ Omi Friend (AggieWorks take-home project)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)

> Enhancing the Omi experience with advanced AI-powered audio analysis

- The hosted version for the take-home can be found here: https://aggie.server.bardia.app
- Demo Account with Loaded Audio Files + Connected to GCP Bucket
    - Email: demo@bardia.app
    - Password: bardia

## 🚀 Project Description

Omi Friend is a companion application designed to work with the Omi device, a continuous audio recorder. The Omi device records everything it hears and sends the audio to your phone. This application aims to improve upon the existing Omi app by providing more accurate transcriptions, better speaker detection, and advanced analysis through various plugins.

The Omi App will automatically output the raw audio files to a GCP bucket (same credentials used in the Omi App under developer mode), where it will get picked up by the webapp after clicking the refresh button in the top right. There is also an option to upload your own auio files through the UI, but that feature is still a little buggy :(.

Key features include:
- Improved transcription using Gladia's Wisper-Zero model
- Advanced conversation analysis using Gladia's 
- Intelligent conversation detection and separation
- Easy playback of recordings
- Music detection and song identification
- Various analysis plugins: sentiment analyzer, bias detector, action item creator, reminder setter, calendar item creator, etc.

## ✨ Current Features

- 🔐 User authentication system
- 📁 GCP bucket integration for audio file retrieval
- 🗣️ Basic audio transcription using Gladia API
- 📊 Simple conversation view with audio player
- 🧩 Plugins marketplace concept (UI only)
- ⚙️ Settings management for GCP and Gladia credentials

## 🚧 Work in Progress

- 🥴 Still a few bugs in the Quick Info section (I'll fix them soon I promise)
- 🤖 AI chat functionality
- 🎯 Improved transcription accuracy
- 👂 Enhanced speaker detection
- 🔍 Advanced search functionality across all transcripts
- 📈 Sentiment analysis and bias detection plugins
- 🎵 Music detection and analysis
- 📱 Mobile responsive design
- 🔗 API integrations with popular communication platforms

## 🛠️ Tech Stack

- **Frontend**: Next.js, React, ShadCN, Tailwind CSS
- **Backend**: Go
- **Database**: MongoDB
- **Cloud Storage**: Google Cloud Platform (GCP)
- **Authentication**: JWT
- **APIs**: Gladia

## 🏃‍♂️ Running the Project

To run this project locally, follow these steps:

1. Clone the repository:
   ```
   git clone https://github.com/TheLickIn13Keys/omi-friend.git
   cd omi-friend
   ```

2. Set up the backend:
   ```
   cd backend
   go mod download
   ```

3. Set up environment variables:
   Create a `.env` file in the backend directory with the following variables:
   ```
   MONGO_URI=your_mongodb_connection_string
   JWT_SECRET=your_jwt_secret
   ```

4. Start the backend server:
   ```
   go run main.go
   ```

5. Set up the frontend:
   ```
   cd ../frontend
   npm install
   ```

6. Start the frontend development server:
   ```
   npm run dev
   ```

7. Open your browser and navigate to `http://localhost:3000`

Note: You'll need to have Go, Node.js, and MongoDB installed on your system.

## 🤝 Contributing

All contributions are welcome! No contributions guide at the moment!

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgements

- [Omi](https://omi.audio/) for the inspiration and raw audio data
- [Gladia](https://www.gladia.io/) for their transcription API (currently used, to be replaced)
- [OpenAI](https://openai.com/) for Whisper and ChatGPT (future implementation)


---

Made with ❤️ by Bardia Anvari for Aggie Works!
