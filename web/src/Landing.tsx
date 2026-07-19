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
  const docs = location.pathname === "/docs"
  const onboardingUrl = `${location.origin}/agent-onboarding/SKILL.md`
  const agentInstruction = `Read and follow ${onboardingUrl}`

  return <main className="landing">
    <nav className="land-nav">
      <a href="/" className="land-brand">mica<span>.</span></a>
      <div className="nav-links">
        <a href="/#product">Product</a>
        <a href="/#workflow">Workflow</a>
        <a href="/docs">Docs</a>
        <a href="/setup">Agent setup</a>
        <a href="https://github.com/pol-cova/mica">GitHub</a>
      </div>
      <a className="button button-dark" href="/workspace">Open workspace</a>
    </nav>

    {setup
      ? <SetupPage onboardingUrl={onboardingUrl} agentInstruction={agentInstruction}/>
      : docs
        ? <DocsPage agentInstruction={agentInstruction}/>
      : <Home onboardingUrl={onboardingUrl} agentInstruction={agentInstruction}/>
    }

    <footer className="land-footer">
      <a href="/" className="land-brand">mica<span>.</span></a>
      <p>Prometheus comparison, incident records, and recovery checks.</p>
      <div><a href="/docs">Docs</a><a href="/setup">Agent setup</a><a href="https://github.com/pol-cova/mica">GitHub</a><a href="https://github.com/pol-cova/mica/blob/main/LICENSE">MIT</a></div>
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
      <SectionLabel index="01 / 04" label="Product"/>
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
      <SectionLabel index="02 / 04" label="Agent workflow"/>
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
      <SectionLabel index="03 / 04" label="Setup"/>
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

    <section className="section docs-section">
      <SectionLabel index="04 / 04" label="Documentation"/>
      <div className="section-heading">
        <h2>Start with the task<br/>you need to finish.</h2>
        <p>Run the demo, follow a human investigation, connect an agent, or configure live Prometheus.</p>
      </div>
      <div className="docs-link-grid">
        <a href="/docs#quick-start"><span>01</span><div><b>Run Mica locally</b><small>Start the stack and create evidence</small></div><i>→</i></a>
        <a href="/docs#human-workflow"><span>02</span><div><b>Use the workspace</b><small>Investigate and verify as a human</small></div><i>→</i></a>
        <a href="/docs#agent-workflow"><span>03</span><div><b>Connect an agent</b><small>Configure MCP and continue an incident</small></div><i>→</i></a>
        <a href="/docs#configuration"><span>04</span><div><b>Use live telemetry</b><small>Catalog and Prometheus configuration</small></div><i>→</i></a>
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
      <article><span>01</span><div><h2>Start Mica</h2><p>Starts the daemon, workspace, Prometheus, and checkout service.</p><pre><code>make demo-up</code><CopyButton value="make demo-up"/></pre></div></article>
      <article><span>02</span><div><h2>Load the setup file</h2><p>Paste this instruction into a compatible coding agent.</p><pre><code>{agentInstruction}</code><CopyButton value={agentInstruction}/></pre><a href={onboardingUrl}>Open SKILL.md →</a></div></article>
      <article><span>03</span><div><h2>Configure MCP manually</h2><p>Use this when the client cannot follow the onboarding skill.</p><pre className="config"><code>{mcpConfig}</code><CopyButton value={mcpConfig}/></pre></div></article>
    </div>
  </section>
}

function DocsPage({ agentInstruction }: { agentInstruction: string }) {
  const startCommand = "make demo-up"
  const triggerCommand = "MICA_DEMO_CONTROL_URL=http://127.0.0.1:8081 go run ./cmd/mica demo trigger n-plus-one"
  return <section className="docs-page grid-surface">
    <aside className="docs-index">
      <span className="page-index">Documentation</span>
      <h1>Use Mica.</h1>
      <p>Commands and workflows for the local demo, human workspace, and MCP agents.</p>
      <nav aria-label="Documentation sections">
        <a href="#quick-start">Quick start</a>
        <a href="#human-workflow">Human workflow</a>
        <a href="#agent-workflow">Agent workflow</a>
        <a href="#tools">MCP tools</a>
        <a href="#configuration">Configuration</a>
      </nav>
    </aside>
    <div className="docs-content">
      <article id="quick-start">
        <DocTitle number="01" title="Quick start" detail="Run the complete local stack and collect a real comparison."/>
        <DocCommand label="Start Mica" value={startCommand}/>
        <p>Open <a href="http://127.0.0.1:8787/workspace">127.0.0.1:8787/workspace</a>. Wait about 30 seconds for a healthy baseline.</p>
        <DocCommand label="Trigger the checkout regression" value={triggerCommand}/>
        <p>Wait about 20 seconds, then select <strong>Compare telemetry</strong>.</p>
      </article>
      <article id="human-workflow">
        <DocTitle number="02" title="Human workflow" detail="The workspace keeps one incident record from detection through recovery."/>
        <ol className="docs-steps">
          <li><span>1</span><div><b>Read the result</b><p>Start with the regression count and the baseline-to-incident values.</p></div></li>
          <li><span>2</span><div><b>Inspect evidence</b><p>Each signal shows baseline, incident, latest value, threshold, query, and collection windows.</p></div></li>
          <li><span>3</span><div><b>Record the investigation</b><p>Save hypotheses, evidence IDs, code changes, tests, and operator notes.</p></div></li>
          <li><span>4</span><div><b>Verify recovery</b><p>After a recorded change and fresh samples, compare the original degraded signals again.</p></div></li>
        </ol>
      </article>
      <article id="agent-workflow">
        <DocTitle number="03" title="Agent workflow" detail="Agents use the same incident record through typed MCP tools."/>
        <DocCommand label="Give this to a compatible agent" value={agentInstruction}/>
        <ol className="docs-steps compact">
          <li><span>1</span><div><b>Verify MCP</b><p>The onboarding file checks health, service context, and the available tools.</p></div></li>
          <li><span>2</span><div><b>Open the Agent tab</b><p>Copy the incident-specific handoff. It includes the service, incident ID, evidence IDs, and next skill.</p></div></li>
          <li><span>3</span><div><b>Watch Activity</b><p>Agent tool calls and saved records appear in the workspace timeline.</p></div></li>
        </ol>
      </article>
      <article id="tools">
        <DocTitle number="04" title="MCP tools" detail="Task-level operations without deployment or infrastructure access."/>
        <div className="tool-table">
          <div><b>Read</b><code>get_service_context</code><code>inspect_service</code><code>find_correlations</code></div>
          <div><b>Record</b><code>record_skill_run</code><code>record_hypothesis</code><code>record_change</code></div>
          <div><b>Act safely</b><code>propose_action</code><code>prepare_incident_update</code><code>verify_recovery</code></div>
        </div>
      </article>
      <article id="configuration">
        <DocTitle number="05" title="Configuration" detail="Keep credentials in the daemon or MCP client environment."/>
        <div className="config-table">
          <div><code>MICA_PROMETHEUS_URL</code><span>Prometheus HTTP API base URL</span></div>
          <div><code>MICA_SERVICE_CATALOG</code><span>Path to the service catalog JSON file</span></div>
          <div><code>MICA_PROMETHEUS_BEARER_TOKEN</code><span>Bearer authentication for Prometheus</span></div>
          <div><code>MICA_PROMETHEUS_BASIC_USER</code><span>Basic authentication user</span></div>
          <div><code>MICA_PROMETHEUS_BASIC_PASSWORD</code><span>Basic authentication password</span></div>
        </div>
        <p>For architecture, validation, and contribution details, use the <a href="https://github.com/pol-cova/mica/tree/main/docs">repository documentation</a>.</p>
      </article>
    </div>
  </section>
}

function DocTitle({ number, title, detail }: { number: string; title: string; detail: string }) {
  return <header className="doc-title"><span>{number}</span><div><h2>{title}</h2><p>{detail}</p></div></header>
}

function DocCommand({ label, value }: { label: string; value: string }) {
  return <div className="doc-command"><span>{label}</span><pre><code>{value}</code><CopyButton value={value}/></pre></div>
}

function ProductWindow() {
  return <div className="product-window">
    <header className="window-bar"><div><i/><i/><i/></div><span>127.0.0.1:8787/workspace</span><b><i/> Prometheus connected</b></header>
    <img className="workspace-shot" src="/product/workspace.png" alt="Mica workspace showing a real checkout-service Prometheus comparison"/>
  </div>
}

function SectionLabel({ index, label }: { index: string; label: string }) {
  return <div className="section-label"><span>[ {index} ]</span><i/> <b>{label}</b></div>
}

function Capability({ number, title, text, detail }: { number: string; title: string; text: string; detail: string }) {
  return <article className="capability"><span>[ {number} ]</span><div className="capability-icon"><i/><i/><i/></div><h3>{title}</h3><p>{text}</p><small>{detail}</small></article>
}
