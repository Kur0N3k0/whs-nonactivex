// 서버가 제공하는 JS: 브라우저 → 로컬 .NET 클라이언트 통신 브리지
(() => {
  async function tryFetch(url, opt = {}) {
    try { return await fetch(url, opt); } catch { return null; }
  }

  const DEFAULT_PORTS = [4589];

  class Bridge {
    constructor(token) {
      this.token = token;
      this.baseURL = null; // e.g., http://127.0.0.1:4589
    }
    async probe() {
      for (const host of ["127.0.0.1", "localhost"]) {
        for (const p of DEFAULT_PORTS) {
          const base = `http://${host}:${p}`;
          // PNA 대응: 프리플라이트 허용 필요(서버측 처리)
          const r = await tryFetch(`${base}/health`, { method: "GET", mode: "cors" });
          if (r && r.ok) { this.baseURL = base; return true; }
        }
      }
      return false;
    }
    async downloadFromServer(serverPath, saveAs /* optional */) {
      if (!this.baseURL) throw new Error("client not connected");
      const url = new URL(serverPath, window.location.origin).toString();
      const body = { url, saveAs: saveAs || null };
      const r = await fetch(`${this.baseURL}/v1/download`, {
        method: "POST",
        mode: "cors",
        headers: {
          "Content-Type": "application/json",
          "X-Client-Token": this.token
        },
        body: JSON.stringify(body)
      });
      return await r.json();
    }
    async uploadLocalFile(localPath) {
      if (!this.baseURL) throw new Error("client not connected");
      const body = { path: localPath, uploadUrl: new URL("/api/upload", window.location.origin).toString() };
      const r = await fetch(`${this.baseURL}/v1/upload`, {
        method: "POST",
        mode: "cors",
        headers: {
          "Content-Type": "application/json",
          "X-Client-Token": this.token
        },
        body: JSON.stringify(body)
      });
      return await r.json();
    }
    static async init({ token }) {
      const b = new Bridge(token);
      await b.probe();
      return b;
    }
  }

  window.ClientBridge = Bridge;
})();
