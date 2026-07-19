import "./Landing.css"

const config=`{
  "mcpServers": {
    "mica": {
      "command": "go",
      "args": ["run", "./cmd/mica", "mcp"]
    }
  }
}`

export default function Landing(){
  const setup=location.pathname==="/setup"
  return <main className="landing"><nav><a href="/" className="land-brand">mica<span>.</span></a><div><a href="#loop">How it works</a><a href="/setup">Agent setup</a><a className="land-open" href="/workspace">Open workspace</a></div></nav>{setup?<section className="setup"><p>Agent setup</p><h1>Connect code<br/>to <em>evidence.</em></h1><span>Run Mica locally, then add its typed MCP server to your agent.</span><div className="setup-card"><b>1. Start the stack</b><code>docker compose up --build</code><b>2. Add Mica to your MCP config</b><pre>{config}</pre><b>3. Start an investigation</b><code>Investigate the checkout regression and verify recovery.</code></div></section>:<><section className="land-hero"><div><p>Local production engineering</p><h1>Fix the code.<br/><em>Check</em> the system.</h1><span>Prometheus evidence, an incident record, and a typed MCP server for production work that can be checked after the change.</span><section className="actions"><a className="land-open" href="/workspace">Open workspace</a><a href="/setup">Set up an agent →</a></section><small>Go daemon · React workspace · MCP over stdio</small></div><article className="live-card"><header><i/> checkout <b>prometheus</b></header><p>Current comparison</p><h2>2 signals changed</h2><div><Metric n="p95 latency" v="+411%"/><Metric n="DB queries" v="+352%"/><Metric n="Error rate" v="8.3%"/></div><footer>Baseline: 10:00–10:15 · Current: 10:20–10:35</footer></article></section><section id="loop" className="land-loop"><p>The loop</p><h2>Observe. Change. Verify.</h2><div><Step n="01" t="Compare" d="Measure the current window against a healthy baseline."/><Step n="02" t="Investigate" d="Give the agent context, evidence IDs, and boundaries."/><Step n="03" t="Verify" d="Check the same signals after the change."/></div></section><section className="land-end"><p>For agents</p><h2>Operational context without raw infrastructure access.</h2><a className="land-open" href="/setup">Agent setup →</a></section></>}<footer className="land-footer">Mica · local daemon · <a href="https://github.com/pol-cova/mica">GitHub</a></footer></main>}
function Metric({n,v}:{n:string;v:string}){return <span><small>{n}</small><b>{v}</b></span>}
function Step({n,t,d}:{n:string;t:string;d:string}){return <article><small>{n}</small><b>{t}</b><span>{d}</span></article>}
