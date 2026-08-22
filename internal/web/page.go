package web

import "net/http"

const Index = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>lichen-atlas</title>
<style>body{font:16px system-ui;margin:2rem;max-width:56rem}code{background:#eef2f0;padding:.15rem .3rem}button{padding:.45rem .8rem}</style></head>
<body><h1>lichen-atlas field desk</h1><p>Sample intake and survey health.</p>
<button id="health">Check service</button><pre id="out"></pre>
<script>document.querySelector('#health').onclick=async()=>{const r=await fetch('/healthz');document.querySelector('#out').textContent=await r.text()}</script>
</body></html>`

func ServeIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(Index))
}
