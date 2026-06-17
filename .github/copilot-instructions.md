goDev - Technical Implementation Prompt

Role: Act as a Senior Systems Engineer, Go Developer, and React/Tailwind Expert.

Context: We are building "goDev", an open-source, blazing-fast local PHP development environment manager (a Laragon / Laravel Herd alternative). The stack is Wails (Go Backend + React/Tailwind Frontend). The frontend UI is already designed. We now need to implement the core functional logic step-by-step, ensuring native OS binaries are orchestrated without Docker.
Crucially, the application must include inbuilt/bundled default versions of Apache, Nginx, MySQL, PHP, and phpMyAdmin so it works immediately out of the box.

Strict Guidelines:

Error Handling: Every Go function must return errors to the React frontend. Do not silently fail. If a port (e.g., 80) is in use, send a specific error message to the UI.

Inbuilt Yet Configurable: The app must default to its bundled binaries. However, do not hardcode these binary paths in the Go logic. Read from a config.yaml or config.json so users can easily swap, configure, or upgrade Apache/Nginx/PHP/MySQL versions by simply dropping new binaries into a localized /tools folder and updating the config.

Cross-Platform: Write Go code that handles Windows (.exe, explorer) and macOS/Linux (open, xdg-open) gracefully where applicable, though prioritize Windows if a choice must be made.

Step-by-Step Execution: Do not give me all the code at once. Implement this sequentially, one phase at a time. Ask for my approval before moving to the next phase.

Phase 1: Configuration & Service Orchestration (The "Start All" Engine)

Default State: The UI should default to "Start All".

Configuration Loader: Create a Go struct that reads config.yaml (defining paths to the bundled nginx.exe, httpd.exe, mysqld.exe, php-cgi.exe and their ports).

Wails Bindings: Create Go methods StartAllServices() and StopAllServices() and expose them to Wails.

Process Execution: Use Go's os/exec to launch the web server, PHP-FPM/CGI, and MySQL as detached, background processes from their inbuilt directories. Capture their PIDs.

State Sync: Return a boolean or state object to React so the UI updates the status dots (Green = Running) and the button text changes to "Stop All".

Phase 2: Auto Virtual Hosts & Local Domains (.test)

Directory Watcher: When services start, scan the /www or /public folder.

VHost Generation: For every folder (e.g., laravel), auto-generate an Nginx or Apache config block setting the server name to laravel.test (The .test suffix must be a variable fetched from user preferences/config).

Hosts File / DNS: Inject 127.0.0.1 laravel.test into the OS hosts file (requires elevated privileges handling) OR implement a lightweight local DNS server in Go to route \*.test.

HTTPS/SSL: Integrate mkcert logic. When generating the vhost, automatically generate a local trusted SSL certificate so the project is accessible via https://laravel.test. Reload the web server gracefully.

Phase 3: Database & phpMyAdmin Integration

UI Updates: In the React UI, clicking the "Database" menu item should display detailed MySQL info (Current inbuilt version running, uptime, port 3306 status).

phpMyAdmin Launcher: Create a Wails-bound Go function LaunchDatabaseManager(). This should:

Ensure the bundled phpMyAdmin is situated in a localized /tools/phpmyadmin folder.

Ensure an Nginx/Apache vhost exists for it (e.g., http://localhost/phpmyadmin or http://phpmyadmin.test).

Open the user's default web browser to that specific URL using Go.

Phase 4: Terminal Integration

Wails Binding: Create a Go function OpenTerminal(path string).

Terminal Detection: When the user clicks "Terminal" in the React UI, trigger this function with the path to the /www folder.

Execution: The Go code should detect the OS and launch a modern terminal.

Windows: Try to launch Windows Terminal (wt.exe -d path), fallback to PowerShell or cmd.exe /K "cd /d path". Optionally support detecting Cmder.

macOS: Launch Terminal.app or iTerm2 scoped to the directory.

Phase 5: Root Folder / Explorer Integration

Wails Binding: Create a Go function OpenRootDirectory().

Execution: When "Root" is clicked in the React UI, the Go backend should open the /www directory in the native file explorer.

Windows: exec.Command("explorer", path)

macOS: exec.Command("open", path)

Linux: exec.Command("xdg-open", path)

Agent: Please acknowledge these requirements. Once acknowledged, begin with Phase 1 by providing the Go struct for configuration and the Go os/exec logic for starting the inbuilt web server and database. Do not proceed to Phase 2 until Phase 1 is complete and functional.
