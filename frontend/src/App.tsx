import { useState, useEffect, useRef } from "react";
import { Quit } from "../wailsjs/runtime/runtime";
import { MinimizeToTray, StartAllServices, StopAllServices } from "../wailsjs/go/main/App";

type ServiceStatus = {
  running: boolean;
  message?: string;
  activeWebServer?: string;
  apachePid?: number;
  nginxPid?: number;
  mysqlPid?: number;
  phpPid?: number;
};

function App() {
  const [isRunning, setIsRunning] = useState(false);
  const [statusMessage, setStatusMessage] = useState("Start All ready");
  const [isBusy, setIsBusy] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // Close context menu when clicking anywhere outside of it
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setMenuOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [menuRef]);

  const handleToggleServices = async () => {
    if (isBusy) {
      return;
    }

    setIsBusy(true);
    try {
      const result: ServiceStatus = isRunning
        ? await StopAllServices()
        : await StartAllServices();

      setIsRunning(result.running);
      setStatusMessage(
        result.message ??
          (result.running ? "Services running" : "Services stopped"),
      );
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      setStatusMessage(message);
    } finally {
      setIsBusy(false);
    }
  };

  return (
    <div
      className="flex h-screen w-screen bg-slate-50 font-sans text-gray-800 overflow-hidden"
      style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}
    >
      {/* --- SIDEBAR --- */}
      <aside className="w-[220px] bg-white border-r border-gray-200 flex flex-col z-10">
        {/* Logo Area (Draggable) */}
        <div
          className="h-[120px] flex items-center justify-center p-5"
          style={{ "--wails-draggable": "drag" } as React.CSSProperties}
        >
          <svg viewBox="0 0 100 100" className="w-[100px] h-[100px]">
            <path
              d="M50 10 A 40 40 0 1 0 90 50"
              fill="none"
              stroke="#0ea5e9"
              strokeWidth="12"
              strokeLinecap="round"
            />
            <circle cx="50" cy="50" r="15" fill="#38bdf8" />
            <path
              d="M50 50 Q 80 20 95 40"
              fill="none"
              stroke="#0ea5e9"
              strokeWidth="8"
              strokeLinecap="round"
            />
          </svg>
        </div>

        {/* --- SIDEBAR NAVIGATION --- */}
        <nav className="flex-1 flex flex-col py-2.5">
          {/* Active Menu Item with Context Menu */}
          <div
            className="flex items-center px-6 py-3 text-sky-600 bg-sky-50 border-l-4 border-sky-600 cursor-pointer text-[14.5px] font-medium relative"
            onClick={() => setMenuOpen(!menuOpen)}
            ref={menuRef}
          >
            <i className="fa-solid fa-house w-6 text-[16px] mr-3 text-center text-sky-600"></i>
            <span>Menu</span>

            {/* The Context Menu Popup */}
            {menuOpen && (
              <div className="absolute top-full left-full -mt-[30px] -ml-1 bg-white border border-gray-200 rounded-md shadow-lg w-[200px] py-1.5 z-[100]">
                <div className="px-4 py-2 text-[13px] text-gray-700 flex justify-between items-center hover:bg-gray-100 cursor-pointer">
                  <span>CyberEnv</span>{" "}
                  <i className="fa-solid fa-caret-right text-[10px] text-gray-400"></i>
                </div>
                <div className="px-4 py-2 text-[13px] text-gray-700 flex justify-between items-center hover:bg-gray-100 cursor-pointer">
                  <span>
                    <i className="fa-solid fa-arrow-up-right-from-square mr-2 text-[12px]"></i>{" "}
                    www
                  </span>
                  <i className="fa-solid fa-caret-right text-[10px] text-gray-400"></i>
                </div>
                <div className="px-4 py-2 text-[13px] text-gray-700 flex justify-between items-center hover:bg-gray-100 cursor-pointer">
                  <span>Quick app</span>{" "}
                  <i className="fa-solid fa-caret-right text-[10px] text-gray-400"></i>
                </div>
                <div className="px-4 py-2 text-[13px] text-gray-700 flex justify-between items-center hover:bg-gray-100 cursor-pointer">
                  <span>Tools</span>{" "}
                  <i className="fa-solid fa-caret-right text-[10px] text-gray-400"></i>
                </div>

                <div className="h-px bg-gray-200 my-1"></div>

                <div className="px-4 py-2 text-[13px] text-gray-700 flex justify-between items-center hover:bg-gray-100 cursor-pointer">
                  <span>PHP</span>{" "}
                  <i className="fa-solid fa-caret-right text-[10px] text-gray-400"></i>
                </div>
                <div className="px-4 py-2 text-[13px] text-gray-700 flex justify-between items-center hover:bg-gray-100 cursor-pointer">
                  <span>Apache</span>{" "}
                  <i className="fa-solid fa-caret-right text-[10px] text-gray-400"></i>
                </div>
                <div className="px-4 py-2 text-[13px] text-gray-700 flex justify-between items-center hover:bg-gray-100 cursor-pointer">
                  <span>MySQL</span>{" "}
                  <i className="fa-solid fa-caret-right text-[10px] text-gray-400"></i>
                </div>

                <div className="h-px bg-gray-200 my-1"></div>

                <div
                  className="px-4 py-2 text-[13px] text-gray-700 flex justify-between items-center hover:bg-gray-100 cursor-pointer"
                  onClick={handleToggleServices}
                >
                  <span>
                    <i
                      className={`fa-solid ${isRunning ? "fa-stop text-red-500" : "fa-play text-emerald-500"} mr-2`}
                    ></i>
                    {isRunning ? "Stop" : "Start"}
                  </span>
                </div>
                <div className="px-4 py-2 text-[13px] text-gray-700 flex justify-between items-center hover:bg-gray-100 cursor-pointer">
                  <span>Preferences...</span>
                </div>
                <div className="px-4 py-2 text-[13px] text-gray-700 flex justify-between items-center hover:bg-gray-100 cursor-pointer">
                  <span onClick={() => void MinimizeToTray()}>Minimize</span>
                </div>
                <div
                  className="px-4 py-2 text-[13px] text-gray-700 flex justify-between items-center hover:bg-gray-100 cursor-pointer"
                  onClick={() => void Quit()}
                >
                  <span>Quit</span>
                </div>
              </div>
            )}
          </div>

          {/* Static Nav Items */}
          <div className="flex items-center px-6 py-3 text-gray-500 hover:bg-gray-50 border-l-4 border-transparent cursor-pointer text-[14.5px] font-medium transition-colors">
            <i className="fa-solid fa-globe w-6 text-[16px] mr-3 text-center text-gray-400"></i>
            <span>Web</span>
          </div>
          <div className="flex items-center px-6 py-3 text-gray-500 hover:bg-gray-50 border-l-4 border-transparent cursor-pointer text-[14.5px] font-medium transition-colors">
            <i className="fa-solid fa-database w-6 text-[16px] mr-3 text-center text-gray-400"></i>
            <span>Database</span>
          </div>
          <div className="flex items-center px-6 py-3 text-gray-500 hover:bg-gray-50 border-l-4 border-transparent cursor-pointer text-[14.5px] font-medium transition-colors">
            <i className="fa-solid fa-terminal w-6 text-[16px] mr-3 text-center text-gray-400"></i>
            <span>Terminal</span>
          </div>
          <div className="flex items-center px-6 py-3 text-gray-500 hover:bg-gray-50 border-l-4 border-transparent cursor-pointer text-[14.5px] font-medium transition-colors">
            <i className="fa-regular fa-folder-open w-6 text-[16px] mr-3 text-center text-gray-400"></i>
            <span>Root</span>
          </div>
        </nav>

        {/* Sidebar Footer (Toggle Button) */}
        <div className="p-5 border-t border-gray-200">
          <button
            className={`flex items-center font-medium text-[14.5px] transition-colors focus:outline-none ${isRunning ? "text-sky-600" : "text-gray-600 hover:text-sky-600"} ${isBusy ? "opacity-60" : ""}`}
            onClick={handleToggleServices}
            disabled={isBusy}
          >
            <i
              className={`fa-regular ${isRunning ? "fa-square" : "fa-circle-play"} text-[18px] mr-3`}
            ></i>
            <span>{isRunning ? "Stop All" : "Start All"}</span>
          </button>
        </div>
      </aside>

      {/* --- MAIN CONTENT --- */}
      <main className="flex-1 bg-slate-50 p-10 flex flex-col">
        {/* Header (Draggable background) */}
        <header
          className="flex justify-between items-center mb-8"
          style={{ "--wails-draggable": "drag" } as React.CSSProperties}
        >
          <h1 className="text-2xl font-semibold text-gray-800">Services</h1>

          <div
            className="flex items-center gap-4"
            style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}
          >
            <button className="flex items-center gap-2 text-sm text-gray-500 hover:text-sky-600 transition-colors focus:outline-none">
              <i className="fa-solid fa-rotate"></i> Reload
            </button>
            <button className="text-[18px] text-gray-500 hover:text-sky-600 transition-colors focus:outline-none">
              <i className="fa-regular fa-circle-question"></i>
            </button>
            <button className="text-[18px] text-gray-500 hover:text-sky-600 transition-colors focus:outline-none">
              <i className="fa-solid fa-gear"></i>
            </button>
          </div>
        </header>

        <div className="mb-6 rounded-lg border border-sky-100 bg-sky-50 px-4 py-3 text-sm text-sky-700 shadow-sm">
          {statusMessage}
        </div>

        {/* Services List */}
        <div className="flex flex-col gap-4">
          {/* Apache Service Card */}
          <div className="bg-white border border-gray-100 rounded-xl px-6 py-5 flex justify-between items-center shadow-sm hover:shadow-md transition-shadow">
            <div className="flex items-center">
              <div
                className={`w-3 h-3 rounded-full mr-5 ${isRunning ? "bg-emerald-500 shadow-[0_0_6px_rgba(16,185,129,0.4)]" : "bg-gray-300"}`}
              ></div>
              <div>
                <h2 className="text-base font-semibold text-gray-900 mb-1">
                  Apache
                </h2>
                <p className="text-[13px] text-gray-500">
                  httpd-2.4.62-win64-VS17
                </p>
              </div>
            </div>
            <div className="flex items-center gap-6">
              <span className="text-[13.5px] text-gray-600 font-medium w-[60px] text-right">
                80/443
              </span>
              {isRunning ? (
                <span className="px-3 py-1 rounded-full text-xs font-semibold w-[75px] text-center bg-emerald-50 text-emerald-600">
                  Running
                </span>
              ) : (
                <span className="px-3 py-1 rounded-full text-xs font-semibold w-[75px] text-center bg-gray-100 text-gray-500">
                  Stopped
                </span>
              )}
              <button className="text-gray-400 hover:text-gray-600 p-1 focus:outline-none">
                <i className="fa-solid fa-ellipsis-vertical text-lg"></i>
              </button>
            </div>
          </div>

          {/* MySQL Service Card */}
          <div className="bg-white border border-gray-100 rounded-xl px-6 py-5 flex justify-between items-center shadow-sm hover:shadow-md transition-shadow">
            <div className="flex items-center">
              <div
                className={`w-3 h-3 rounded-full mr-5 ${isRunning ? "bg-emerald-500 shadow-[0_0_6px_rgba(16,185,129,0.4)]" : "bg-gray-300"}`}
              ></div>
              <div>
                <h2 className="text-base font-semibold text-gray-900 mb-1">
                  MySQL
                </h2>
                <p className="text-[13px] text-gray-500">mysql-8.0.30-winx64</p>
              </div>
            </div>
            <div className="flex items-center gap-6">
              <span className="text-[13.5px] text-gray-600 font-medium w-[60px] text-right">
                3306
              </span>
              {isRunning ? (
                <span className="px-3 py-1 rounded-full text-xs font-semibold w-[75px] text-center bg-emerald-50 text-emerald-600">
                  Running
                </span>
              ) : (
                <span className="px-3 py-1 rounded-full text-xs font-semibold w-[75px] text-center bg-gray-100 text-gray-500">
                  Stopped
                </span>
              )}
              <button className="text-gray-400 hover:text-gray-600 p-1 focus:outline-none">
                <i className="fa-solid fa-ellipsis-vertical text-lg"></i>
              </button>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

export default App;
