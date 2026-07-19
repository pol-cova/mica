import { useState } from "react"
import "./Landing.css"
import "./LandingV2.css"

const mcpConfig = `{
  "mcpServers": {
    "mica": {
      "command": "go",
      "args": ["run", "./cmd/mica", "mcp"]
    }
  }
}`

function CopyButton({ value, label = "Copy" }: { value: string; label?: string }) {
  const [copied, setCopied] = useState(false)

  async function copy() {
    await navigator.clipboard.writeText(value)
    setCopied(true)
    window.setTimeout(() => setCopied(false), 1600)
  }

  return <button className="copy-button" onClick={copy}>{copied ? "Copied" : label}</button>
}

export default function Landing() {
  const setup = location.pathname === "/setup"
  const onboardingUrl = `${location.origin}/agent-onboarding/SKILL.md`
  const agentInstruction = `Read and follow ${onboardingUrl}`

  return <main className="landing">
    <nav className="land-nav">
      <a href="/" className="land-brand">mica<span>.</span></a>
      <div className="nav-links">
        <a href="/#product">Product</a>
        <a href="/#workflow">Workflow</a>
        <a href="/setup">Agent setup</a>
        <a href="https://github.com/pol-cova/mica">GitHub</a>
      </div>
      <a className="button button-dark" href="/workspace">Open workspace</a>
    </nav>

    {setup
      ? <SetupPage onboardingUrl={onboardingUrl} agentInstruction={agentInstruction}/>
      : <Home onboardingUrl={onboardingUrl} agentInstruction={agentInstruction}/>
    }

    <footer className="land-footer">
      <a href="/" className="land-brand">mica<span>.</span></a>
      <p>Prometheus comparison, incident records, and recovery checks.</p>
      <div><a href="/setup">Agent setup</a><a href="https://github.com/pol-cova/mica">GitHub</a><a href="https://github.com/pol-cova/mica/blob/main/LICENSE">MIT</a></div>
    </footer>
  </main>
}

function Home({ onboardingUrl, agentInstruction }: { onboardingUrl: string; agentInstruction: string }) {
  return <>
    <section className="hero grid-surface">
      <div className="hero-copy">
        <h1>Investigate production<br/>regressions <em>with evidence.</em></h1>
        <p className="hero-summary">Compare Prometheus windows, record what changed, and check the same signals after the fix.</p>
        <div className="hero-actions">
          <a className="button button-blue" href="/workspace">Open workspace</a>
          <a className="button button-light" href="/setup">Connect an agent <span>→</span></a>
        </div>
        <div className="agent-command">
          <span className="command-mark">›_</span>
          <code>{agentInstruction}</code>
          <CopyButton value={agentInstruction}/>
        </div>
      </div>
      <ProductWindow/>
    </section>

    <section id="product" className="section section-product">
      <SectionLabel index="01 / 03" label="Product"/>
      <div className="section-heading">
        <h2>One incident record.<br/>Three clear steps.</h2>
        <p>Evidence, decisions, changes, and recovery results stay attached to the same incident.</p>
      </div>
      <div className="capability-grid">
        <Capability number="01" title="Compare" text="Run the same Prometheus queries over a healthy baseline and the incident window." detail="Baseline → incident"/>
        <Capability number="02" title="Record" text="Store hypotheses, evidence references, code changes, and human decisions." detail="SQLite incident record"/>
        <Capability number="03" title="Verify" text="Run fresh queries after the change and compare the affected signals again." detail="Incident → now"/>
      </div>
    </section>

    <section id="workflow" className="section workflow-section">
      <SectionLabel index="02 / 03" label="Agent workflow"/>
      <div className="workflow-layout">
        <div>
          <h2>Typed tools for the investigation.</h2>
          <p>Mica exposes service context, signal comparison, record updates, and recovery checks through MCP. It does not expose deployment or infrastructure controls.</p>
          <a href="/setup">Agent setup →</a>
        </div>
        <ol className="workflow-list">
          <li><span>01</span><div><b>Get service context</b><small>repository, runtime, dependencies, runbooks</small></div><code>get_service_context</code></li>
          <li><span>02</span><div><b>Compare signals</b><small>healthy baseline against incident window</small></div><code>detect_regressions</code></li>
          <li><span>03</span><div><b>Record the change</b><small>commit, files, rationale, evidence IDs</small></div><code>record_change</code></li>
          <li><span>04</span><div><b>Verify recovery</b><small>fresh readings for affected signals</small></div><code>verify_recovery</code></li>
        </ol>
      </div>
    </section>

    <section className="section agent-section">
      <SectionLabel index="03 / 03" label="Setup"/>
      <div className="agent-panel">
        <div className="agent-panel-copy">
          <h2>Connect your agent.</h2>
          <p>Paste one instruction. The setup file checks the daemon, configures MCP, verifies the tool list, and selects a task skill.</p>
          <a className="text-link" href={onboardingUrl}>Read SKILL.md →</a>
        </div>
        <div className="terminal-card">
          <header><i/><i/><i/><span>agent setup</span></header>
          <div><span className="prompt">$</span><code>{agentInstruction}</code></div>
          <CopyButton value={agentInstruction} label="Copy instruction"/>
        </div>
      </div>
    </section>
  </>
}

function SetupPage({ onboardingUrl, agentInstruction }: { onboardingUrl: string; agentInstruction: string }) {
  return <section className="setup-page grid-surface">
    <div className="setup-intro">
      <span className="page-index">01 — Agent setup</span>
      <h1>Connect Mica<br/>to your agent.</h1>
      <p>Start the local stack, load the onboarding file, and verify the connection.</p>
    </div>
    <div className="setup-steps">
      <article><span>01</span><div><h2>Start Mica</h2><p>Starts the daemon, workspace, Prometheus, and checkout service.</p><pre><code>docker compose up --build</code><CopyButton value="docker compose up --build"/></pre></div></article>
      <article><span>02</span><div><h2>Load the setup file</h2><p>Paste this instruction into a compatible coding agent.</p><pre><code>{agentInstruction}</code><CopyButton value={agentInstruction}/></pre><a href={onboardingUrl}>Open SKILL.md →</a></div></article>
      <article><span>03</span><div><h2>Configure MCP manually</h2><p>Use this when the client cannot follow the onboarding skill.</p><pre className="config"><code>{mcpConfig}</code><CopyButton value={mcpConfig}/></pre></div></article>
    </div>
  </section>
}

function ProductWindow() {
  return <div className="product-window">
    <header className="window-bar"><div><i/><i/><i/></div><span>127.0.0.1:8787/workspace</span><b><i/> Prometheus connected</b></header>
    <div className="window-body">
      <aside><b>checkout</b><span>production</span><nav><i className="active"/>Compare<i/>Investigation<i/>Recovery</nav><small>Incident #80073000</small></aside>
      <section>
        <header><div><small>Incident comparison</small><h3>Checkout latency regression</h3></div><button>Run comparison</button></header>
        <div className="metric-row">
          <Metric name="p95 latency" value="+26%" state="Degraded" tone="warn"/>
          <Metric name="error rate" value="0%" state="No change" tone="calm"/>
          <Metric name="DB queries / request" value="+23%" state="Degraded" tone="warn"/>
        </div>
        <div className="evidence-card">
          <div className="chart-title"><div><small>p95 latency</small><b>47.5 ms</b></div><span>Baseline → Incident → Now</span></div>
          <svg viewBox="0 0 720 150" role="img" aria-label="Latency comparison chart">
            <defs><linearGradient id="area" x1="0" x2="0" y1="0" y2="1"><stop offset="0" stopColor="#1769ff" stopOpacity=".2"/><stop offset="1" stopColor="#1769ff" stopOpacity="0"/></linearGradient></defs>
            <path className="grid-line" d="M0 30H720M0 75H720M0 120H720"/>
            <path className="chart-area" d="M0 112 C80 110 120 105 180 108 S300 90 360 65 S470 44 540 46 S660 42 720 44 L720 150 L0 150Z"/>
            <path className="chart-line" d="M0 112 C80 110 120 105 180 108 S300 90 360 65 S470 44 540 46 S660 42 720 44"/>
            <circle cx="180" cy="108" r="4"/><circle cx="360" cy="65" r="4"/><circle cx="720" cy="44" r="4"/>
          </svg>
          <div className="axis"><span>Baseline</span><span>Incident</span><span>Now</span></div>
        </div>
      </section>
    </div>
  </div>
}

function Metric({ name, value, state, tone }: { name: string; value: string; state: string; tone: string }) {
  return <article className={`metric ${tone}`}><small>{name}</small><b>{value}</b><span>{state}</span></article>
}

function SectionLabel({ index, label }: { index: string; label: string }) {
  return <div className="section-label"><span>[ {index} ]</span><i/> <b>{label}</b></div>
}

function Capability({ number, title, text, detail }: { number: string; title: string; text: string; detail: string }) {
  return <article className="capability"><span>[ {number} ]</span><div className="capability-icon"><i/><i/><i/></div><h3>{title}</h3><p>{text}</p><small>{detail}</small></article>
}
