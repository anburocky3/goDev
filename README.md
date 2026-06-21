# goDev - Local Development Environment Manager

![Wails](https://img.shields.io/badge/Wails-v2.12.0-blue)
![Go](https://img.shields.io/badge/Go-1.26.4-brightgreen)
![React](https://img.shields.io/badge/React-v18.2.0-61dafb)
![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)

A powerful desktop application for managing local development services (Apache, Nginx, MySQL, PHP) with an intuitive system tray interface. Built with Go and Wails for cross-platform support on Windows, macOS, and Linux.

**Author:** Anbuselvan Annamalai

## 🌟 Features

- **Service Management**: Start, stop, and monitor Apache, Nginx, MySQL, and PHP services
- **System Tray Integration**: Minimize to system tray for background operation
- **Cross-Platform**: Supports Windows, macOS, and Linux
- **Modern UI**: Clean interface built with React and Tailwind CSS
- **Configuration Management**: YAML-based configuration for easy customization
- **Auto-start Services**: Start all configured services with a single click

## 📦 Tech Stack

### Backend

- **[Wails](https://wails.io/)** v2.12.0 - Cross-platform desktop app framework
- **[Go](https://golang.org/)** 1.25.0 - Programming language
- **[systray](https://github.com/energye/systray)** - System tray integration

### Frontend

- **React** v18.2.0 - UI library
- **TypeScript** v6.0.3 - Type-safe JavaScript
- **Vite** v8.0.16 - Build tool and dev server
- **Tailwind CSS** v4.3.1 - Utility-first CSS framework
- **Bun** - Fast JavaScript package manager and runtime

## 🚀 Prerequisites

Before you begin, ensure you have the following installed:

- [Go](https://golang.org/dl/) (version 1.25.0 or later)
- [Bun](https://bun.sh/) - Fast JavaScript runtime and package manager
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) - Wails command-line tools

## 📋 Installation

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/goDev.git
cd goDev
```

### 2. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install frontend dependencies using Bun
cd frontend
bun install
cd ..
```

### 3. Configure Your Services

Edit the `config.yaml` file to configure your services:

```yaml
services:
  apache:
    enabled: true
    port: 80
    document_root: "/path/to/your/www"

  nginx:
    enabled: false
    port: 8080

  mysql:
    enabled: true
    port: 3306
    data_directory: "/path/to/mysql/data"

  php:
    enabled: true
    version: "8.2"
```

## 🛠️ Development

### Running in Development Mode

```bash
# Start the Wails dev server with hot-reload
wails dev
```

This will start both the Go backend and React frontend development servers automatically.

### Project Structure

```
goDev/
├── build/                 # Build output directory
│   ├── bin/              # Compiled binaries
│   ├── darwin/           # macOS specific files
│   └── windows/          # Windows specific files
├── config.yaml           # Configuration file
├── frontend/             # React frontend application
│   ├── src/
│   │   ├── App.tsx       # Main application component
│   │   ├── main.tsx      # Entry point
│   │   └── global.css    # Tailwind CSS imports
│   ├── package.json      # Frontend dependencies
│   └── vite.config.ts    # Vite configuration
├── services.go           # Service management logic
├── app.go               # Main application structure
├── main.go              # Application entry point
├── wails.json           # Wails configuration
├── go.mod               # Go module dependencies
└── tools/               # Service binaries and configurations
    ├── apache/          # Apache service files
    ├── mysql/           # MySQL service files
    ├── nginx/           # Nginx service files
    └── php/             # PHP service files
```

## 🎮 Usage

### Starting Services

1. Launch the goDev application
2. Click "Start All" to launch all configured services
3. Monitor service status in the main window

### Minimizing to System Tray

- Click the minimize button or use the system tray icon
- Services continue running in the background
- Right-click the system tray icon for quick access

### Managing Individual Services

- Access individual service controls from the main window
- Start/stop services independently
- View service logs and status information

## 📦 Building for Production

### Build for Current Platform

```bash
wails build
```

### Build for Specific Platforms

```bash
# Windows
wails build -platform windows/amd64

# macOS Intel
wails build -platform darwin/amd64

# macOS Apple Silicon
wails build -platform darwin/arm64

# Linux
wails build -platform linux/amd64
```

### Building with Upx Compression (Optional)

```bash
wails build -upx
```

## 🔧 Configuration

The `config.yaml` file contains all service configurations:

- **Service Settings**: Enable/disable individual services
- **Port Configuration**: Customize ports for each service
- **Path Mappings**: Configure document roots and data directories
- **Global Settings**: Application-wide preferences

## 📝 Scripts Overview

### Frontend Scripts (in `frontend/package.json`)

```bash
bun run dev          # Start development server with hot-reload
bun run build        # Build for production
bun run preview      # Preview production build locally
```

## 🤝 Contributing

Contributions are welcome! Please see our [Contributing Guide](CONTRIBUTING.md) for details on:

- Code of Conduct
- Development setup
- Coding guidelines
- How to submit pull requests
- Commit message conventions

## 📄 License

This project is licensed under the [MIT License](LICENSE) - see the license file for details.

Copyright (c) 2026 Anbuselvan Annamalai

## 👥 Authors

- **Anbuselvan Annamalai** - *Initial work* - [15264938+anburocky3@users.noreply.github.com](mailto:15264938+anburocky3@users.noreply.github.com)

## 🙏 Acknowledgments

- [Wails](https://wails.io/) - Cross-platform desktop app framework
- [React](https://reactjs.org/) - UI library
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework
- All [contributors](https://github.com/yourusername/goDev/graphs/contributors) who have helped shape goDev

## 🔗 Links

- [Issue Tracker](https://github.com/yourusername/goDev/issues)
- [Security Policy](SECURITY.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)

### Common Issues

1. **Services won't start**: Ensure the service binaries are properly installed in the `tools/` directory
2. **Port conflicts**: Check that no other applications are using the configured ports
3. **Permission errors**: Run with appropriate permissions or use administrator privileges on Windows
4. **Build failures**: Ensure all dependencies are installed and Go version is compatible

### Logs

- Application logs can be found in the system log (accessible via the application menu)
- Service-specific logs are stored in the `logs/` directory within each service folder

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👤 Author

**Anbuselvan Rocky**

- GitHub: [@anburocky3](https://github.com/anburocky3)
- Email: 15264938+anburocky3@users.noreply.github.com

## 🙏 Acknowledgments

- [Wails Team](https://github.com/wailsapp/wails) for the excellent desktop framework
- React and Tailwind CSS communities for their amazing tools
- The open-source community for all the libraries used in this project

---

**Built with ❤️ using Wails and React**
