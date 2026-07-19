import { useEffect, useMemo, useState } from "react"
import type { FormEvent, ReactNode } from "react"
import { AreaChart } from "@/components/dither-kit/area-chart"
import { Area } from "@/components/dither-kit/area"
import { Tooltip } from "@/components/dither-kit/tooltip"
import { XAxis } from "@/components/dither-kit/x-axis"
import "./App.css"

type Service = {
  id: string
  name: string
  environment: string
  repository: string
  owners: string[]
  dependencies: string[]
  signals: Record<string, string>
  runtime?: string
  framework?: string
  runbooks?: string[]
  recentChanges?: string[]
  source?: { kind: string; refreshedAt: string }
}

type Evidence = {
  id: string
  signal: string
  unit: string
  query: string
  baselineValue: number
  incidentValue: number
  currentValue?: number
  percentDelta: number
  classification: string
}

type Event = { id: string; summary: string; occurredAt: string; actorName?: string }
type Hypothesis = { id: string; summary: string; confidence: string; evidenceIds: string[]; alternatives: string[]; nextStep: string }
type Change = { id: string; summary: string; files: string[]; tests: string[] }
type Proposal = { id: string; summary: string; status: string; riskLevel: string; expectedEffect: string; riskStatement: string; verificationPlan: string; evidenceIds?: string[] }
type Receipt = { destinationId: string; provider: string; status: string; attempts: number; response?: string; deliveredAt?: string }
type Update = { id: string; status: string; body: string; updateType: string; audience: string; redactions: string[]; destinations: Receipt[]; approvedBy?: string }
type Finding = { id: string; severity: string; category: string; assessment: string; recommendation: string; owner?: string; verificationMethod: string; status: string; evidenceRefs: string[] }
type Verification = { status: string; checkedAt: string; signals: Evidence[] }
type Incident = {
  id: string
  serviceId: string
  status: string
  baselineStart: string
  baselineEnd: string
  incidentStart: string
  incidentEnd: string
  evidence: Evidence[]
  timeline: Event[]
  hypotheses: Hypothesis[]
  changes: Change[]
  proposals: Proposal[]
  updates: Update[]
  auditFindings: Finding[]
  verification?: Verification
}
type Destination = { id: string; provider: string }
type Panel = "work" | "recovery" | "context" | "share" | "review"

const iso = (offset: number) => new Date(Date.now() + offset).toISOString()
const label = (value: string) => value.replaceAll("_", " ")
const measure = (value: number, unit: string) => `${value.toLocaleString(undefined, { maximumFractionDigits: 1 })}${unit ? ` ${unit}` : ""}`
const normalizeIncident = (incident: Incident): Incident => ({
  ...incident,
  hypotheses: incident.hypotheses ?? [],
  changes: incident.changes ?? [],
  proposals: incident.proposals ?? [],
  updates: incident.updates ?? [],
  auditFindings: incident.auditFindings ?? [],
  timeline: incident.timeline ?? [],
  evidence: incident.evidence ?? [],
})

async function post<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(path, { method: "POST", headers: { "content-type": "application/json" }, body: JSON.stringify(body) })
  if (!response.ok) throw new Error((await response.json().catch(() => null))?.error ?? "Request failed")
  return response.json()
}

function SignalChart({ item }: { item: Evidence }) {
  const data = useMemo(() => [
    { window: "Baseline", value: item.baselineValue },
    { window: "Incident", value: item.incidentValue },
    { window: "Now", value: item.currentValue ?? item.incidentValue },
  ], [item])

  return <div className="mini-chart" role="img" aria-label={`${item.signal}: ${measure(item.baselineValue, item.unit)} at baseline and ${measure(item.currentValue ?? item.incidentValue, item.unit)} now`}>
    <AreaChart data={data} config={{ value: { label: item.signal, color: "blue" } }} margins={{ left: 8, bottom: 24 }}>
      <XAxis dataKey="window"/>
      <Tooltip labelKey="window" valueFormatter={value => measure(value, item.unit)} variant="frosted-glass"/>
      <Area dataKey="value" variant="gradient"/>
    </AreaChart>
  </div>
}

export default function App() {
  const [services, setServices] = useState<Service[]>([])
  const [service, setService] = useState<Service | null>(null)
  const [incident, setIncident] = useState<Incident | null>(null)
  const [destinations, setDestinations] = useState<Destination[]>([])
  const [chosen, setChosen] = useState<string[]>([])
  const [busy, setBusy] = useState(false)
  const [notice, setNotice] = useState("")
  const [panel, setPanel] = useState<Panel>("work")
  const [updateType, setUpdateType] = useState("investigation_update")
  const [audience, setAudience] = useState("engineering")
  const [approver, setApprover] = useState("")

  const refresh = async (id = incident?.id) => {
    if (id) setIncident(normalizeIncident(await fetch(`/api/incidents/${id}`).then(response => response.json())))
  }

  useEffect(() => {
    Promise.all([
      fetch("/api/services").then(response => response.json()),
      fetch("/api/incidents").then(response => response.json()),
      fetch("/api/communications/destinations").then(response => response.json()),
    ]).then(([serviceList, incidentList, destinationList]) => {
      setServices(serviceList)
      setService(serviceList[0] ?? null)
      setIncident(incidentList[0] ? normalizeIncident(incidentList[0]) : null)
      setDestinations(destinationList)
    }).catch(() => setNotice("Mica is unavailable. Check that the local daemon is running."))
  }, [])

  useEffect(() => {
    if (!incident) return
    const stream = new EventSource(`/api/events?incidentId=${incident.id}`)
    stream.addEventListener("incident-update", () => refresh(incident.id))
    return () => stream.close()
  }, [incident?.id])

  const run = async (work: () => Promise<Incident>, message: string) => {
    setBusy(true)
    setNotice("")
    try {
      setIncident(normalizeIncident(await work()))
      setNotice(message)
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "Request failed")
    } finally {
      setBusy(false)
    }
  }

  const compare = () => service && run(
    () => post("/api/incidents/detect", { serviceId: service.id, baselineStart: iso(-210000), baselineEnd: iso(-120000), incidentStart: iso(-110000), incidentEnd: iso(0) }),
    "Comparison saved.",
  )
  const verify = () => incident && run(
    () => post(`/api/incidents/${incident.id}/verify`, { verificationStart: iso(-90000), verificationEnd: iso(0) }),
    "Recovery check saved.",
  )
  const prepare = () => incident && run(
    () => post(`/api/incidents/${incident.id}/updates`, { updateType, audience, destinationIds: chosen, preparedBy: "Operator" }),
    "Update draft saved.",
  )
  const publish = (update: Update) => {
    if (!approver.trim()) {
      setNotice("Enter the approver name before sending this update.")
      return
    }
    if (incident) run(
      () => post(`/api/incidents/${incident.id}/updates/${update.id}/publish`, { approvedBy: approver.trim() }),
      update.status === "partially_delivered" ? "Failed destinations retried." : "Update sent.",
    )
  }
  const review = (proposal: Proposal, status: "reviewed" | "deferred" | "rejected") => incident && run(
    () => post(`/api/incidents/${incident.id}/proposals/${proposal.id}/review`, { status, reviewer: "Operator", note: "Reviewed in workspace" }),
    `Proposal ${status}.`,
  )
  const submit = (path: string, body: (form: FormData) => unknown, message: string) => (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!incident) return
    const form = event.currentTarget
    run(() => post(`/api/incidents/${incident.id}/${path}`, body(new FormData(form))), message)
    form.reset()
  }
  const copyReport = async () => {
    if (!incident) return
    await navigator.clipboard.writeText(await fetch(`/api/incidents/${incident.id}/report`).then(response => response.text()))
    setNotice("Report copied.")
  }

  const activeService = incident ? services.find(item => item.id === incident.serviceId) ?? service : service

  return <main className="app">
    <Header service={service} services={services} busy={busy} onService={setService} onCompare={compare}/>
    {!incident
      ? <ServiceOverview service={service} busy={busy} onCompare={compare}/>
      : <IncidentWorkspace
          incident={incident}
          service={activeService}
          panel={panel}
          busy={busy}
          notice={notice}
          destinations={destinations}
          chosen={chosen}
          updateType={updateType}
          audience={audience}
          approver={approver}
          onPanel={setPanel}
          onVerify={verify}
          onCopyReport={copyReport}
          onChosen={setChosen}
          onUpdateType={setUpdateType}
          onAudience={setAudience}
          onApprover={setApprover}
          onPrepare={prepare}
          onPublish={publish}
          onReview={review}
          onSubmit={submit}
        />
    }
  </main>
}

function Header({ service, services, busy, onService, onCompare }: {
  service: Service | null
  services: Service[]
  busy: boolean
  onService: (service: Service | null) => void
  onCompare: () => void
}) {
  return <header className="app-header">
    <a href="/" className="brand" aria-label="Mica home">mica<span>.</span></a>
    <label className="service-picker">
      <span>Service</span>
      <select aria-label="Service" value={service?.id ?? ""} onChange={event => onService(services.find(item => item.id === event.target.value) ?? null)}>
        {services.map(item => <option value={item.id} key={item.id}>{item.name} · {item.environment}</option>)}
      </select>
    </label>
    <button className="primary-button" onClick={onCompare} disabled={busy || !service}>{busy ? "Comparing…" : "Compare telemetry"}</button>
    <span className="mode">Local · read-only telemetry</span>
  </header>
}

function ServiceOverview({ service, busy, onCompare }: { service: Service | null; busy: boolean; onCompare: () => void }) {
  if (!service) return <section className="empty-state"><h1>No service configured</h1><p>Add a service to the Mica catalog, then restart the daemon.</p></section>
  const signalNames = Object.keys(service.signals ?? {}).map(label)

  return <section className="service-overview">
    <div className="overview-copy">
      <span className="section-kicker">{service.environment} · {service.source?.kind ?? "configured context"}</span>
      <h1>{service.name}</h1>
      <p>No comparison has been recorded for this service.</p>
      <button className="primary-button" onClick={onCompare} disabled={busy}>{busy ? "Comparing…" : "Compare telemetry"}</button>
    </div>
    <div className="overview-context">
      <ContextItem label="Repository" value={service.repository || "Not configured"}/>
      <ContextItem label="Owner" value={service.owners?.join(", ") || "Not configured"}/>
      <ContextItem label="Runtime" value={[service.runtime, service.framework].filter(Boolean).join(" · ") || "Not configured"}/>
      <ContextItem label="Dependencies" value={service.dependencies?.join(", ") || "None configured"}/>
      <ContextItem label="Signals" value={signalNames.join(", ") || "No signals mapped"}/>
      <ContextItem label="Runbook" value={service.runbooks?.join(", ") || "Not configured"}/>
    </div>
  </section>
}

type WorkspaceProps = {
  incident: Incident
  service: Service | null
  panel: Panel
  busy: boolean
  notice: string
  destinations: Destination[]
  chosen: string[]
  updateType: string
  audience: string
  approver: string
  onPanel: (panel: Panel) => void
  onVerify: () => void
  onCopyReport: () => void
  onChosen: (destinations: string[]) => void
  onUpdateType: (value: string) => void
  onAudience: (value: string) => void
  onApprover: (value: string) => void
  onPrepare: () => void
  onPublish: (update: Update) => void
  onReview: (proposal: Proposal, status: "reviewed" | "deferred" | "rejected") => void
  onSubmit: (path: string, body: (form: FormData) => unknown, message: string) => (event: FormEvent<HTMLFormElement>) => void
}

function IncidentWorkspace(props: WorkspaceProps) {
  const { incident, service, panel, notice } = props
  const degraded = incident.evidence.filter(item => item.classification === "degraded")
  const hasHypothesis = incident.hypotheses.length > 0
  const hasChange = incident.changes.length > 0
  const verified = Boolean(incident.verification)
  const recovered = incident.verification?.status === "recovered"
  const outcome = recovered ? "Recovery verified" : verified ? "Recovery needs attention" : degraded.length ? `${degraded.length} signals changed` : "No required signals changed"
  const measuredSummary = degraded.length
    ? degraded.map(item => `${item.signal} ${item.percentDelta > 0 ? "+" : ""}${item.percentDelta.toFixed(0)}%`).join(" · ")
    : "All required signals are within their configured tolerances."
  const next = !degraded.length ? "No investigation action is required." : !hasHypothesis ? "Record a hypothesis linked to the evidence." : !hasChange ? "Record the code change and tests." : !verified ? "Collect fresh telemetry and verify recovery." : recovered ? "Review the report or prepare an update." : "Review the failed signals and continue the investigation."
  const nextPanel: Panel = !hasHypothesis || !hasChange ? "work" : !verified ? "recovery" : recovered ? "share" : "work"
  const openNextStep = () => {
    props.onPanel(nextPanel)
    requestAnimationFrame(() => document.getElementById("incident-workbench")?.scrollIntoView({ behavior: "smooth", block: "start" }))
  }

  return <>
    <section className="incident-summary">
      <div className="incident-meta">
        <span className={`status-dot ${incident.status}`}/>
        <b>{label(incident.status)}</b>
        <span>#{incident.id.slice(-8)}</span>
        <span>{service?.owners?.join(", ") || "Owner not configured"}</span>
        <time>{new Date(incident.incidentEnd).toLocaleString()}</time>
      </div>
      <div className="decision-row">
        <div>
          <span className="section-kicker">{service?.name ?? incident.serviceId} · {service?.environment ?? "local"}</span>
          <h1>{outcome}</h1>
          <p>{measuredSummary}</p>
        </div>
        <div className="next-card">
          <span>Next</span>
          <p>{next}</p>
          {degraded.length > 0 && <button className="text-button" onClick={openNextStep}>Open next step →</button>}
        </div>
      </div>
      <ProgressSteps incident={incident}/>
    </section>

    <section className="evidence-layout">
      <div className="evidence-main">
        <SectionHeader title="Evidence" detail={`${new Date(incident.baselineStart).toLocaleTimeString()} baseline · ${new Date(incident.incidentEnd).toLocaleTimeString()} incident`}/>
        <div className="chart-grid">
          {incident.evidence.map(item => <EvidenceCard key={item.id} item={item} incident={incident} service={service}/>) }
        </div>
      </div>
      <aside className="activity-card">
        <div className="activity-header">
          <h2>Activity</h2>
          <div><button onClick={props.onCopyReport}>Copy report</button><a href={`/api/incidents/${incident.id}/report`} download>Download</a></div>
        </div>
        <div className="timeline">
          {incident.timeline.slice().reverse().map(event => <div className="event" key={event.id}><i/><span>{event.summary}<small>{event.actorName ?? "Mica"} · {new Date(event.occurredAt).toLocaleTimeString()}</small></span></div>)}
        </div>
      </aside>
    </section>

    <section className="workbench" id="incident-workbench">
      <nav className="panel-tabs" aria-label="Incident workspace sections">
        <PanelTab active={panel === "work"} onClick={() => props.onPanel("work")} label="Investigation" count={incident.hypotheses.length + incident.changes.length}/>
        <PanelTab active={panel === "recovery"} onClick={() => props.onPanel("recovery")} label="Recovery" count={verified ? 1 : 0}/>
        <PanelTab active={panel === "context"} onClick={() => props.onPanel("context")} label="Context"/>
        <PanelTab active={panel === "share"} onClick={() => props.onPanel("share")} label="Share" count={incident.updates.length}/>
        <PanelTab active={panel === "review"} onClick={() => props.onPanel("review")} label="Review" count={incident.proposals.length + incident.auditFindings.length}/>
      </nav>
      <div className="panel-content">
        {panel === "work" && <InvestigationPanel incident={incident} onSubmit={props.onSubmit}/>}
        {panel === "recovery" && <RecoveryPanel incident={incident} busy={props.busy} onVerify={props.onVerify}/>}
        {panel === "context" && <ContextPanel service={service}/>}
        {panel === "share" && <SharePanel {...props}/>}
        {panel === "review" && <ReviewPanel incident={incident} onReview={props.onReview}/>}
      </div>
    </section>

    {notice && <p className="notice" role="status">{notice}</p>}
    <footer className="app-footer">Local incident record · Prometheus comparison windows</footer>
  </>
}

function ProgressSteps({ incident }: { incident: Incident }) {
  const steps = [
    { label: "Evidence", complete: incident.evidence.length > 0 },
    { label: "Diagnosis", complete: incident.hypotheses.length > 0 },
    { label: "Change", complete: incident.changes.length > 0 },
    { label: "Recovery", complete: incident.verification?.status === "recovered" },
  ]
  const firstIncomplete = steps.findIndex(step => !step.complete)
  return <ol className="progress-steps" aria-label="Incident progress">
    {steps.map((step, index) => <li key={step.label} className={step.complete ? "complete" : index === firstIncomplete ? "current" : ""}><span>{index + 1}</span>{step.label}</li>)}
  </ol>
}

function EvidenceCard({ item, incident, service }: { item: Evidence; incident: Incident; service: Service | null }) {
  return <article className="evidence-card">
    <div className="evidence-card-header">
      <div><h3>{item.signal}</h3><span>{item.id}</span></div>
      <div className={`signal-change ${item.classification}`}><b>{item.percentDelta > 0 ? "+" : ""}{item.percentDelta.toFixed(0)}%</b><span>{label(item.classification)}</span></div>
    </div>
    <SignalChart item={item}/>
    <div className="value-row"><span>Baseline <b>{measure(item.baselineValue, item.unit)}</b></span><span>Current <b>{measure(item.currentValue ?? item.incidentValue, item.unit)}</b></span></div>
    <details className="provenance">
      <summary>Query and provenance</summary>
      <dl>
        <div><dt>Service</dt><dd>{service?.name ?? incident.serviceId}</dd></div>
        <div><dt>Source</dt><dd>detect_regressions</dd></div>
        <div><dt>Collected</dt><dd>{new Date(incident.incidentEnd).toLocaleString()}</dd></div>
        <div><dt>Baseline</dt><dd>{new Date(incident.baselineStart).toLocaleString()} to {new Date(incident.baselineEnd).toLocaleString()}</dd></div>
        <div><dt>Incident</dt><dd>{new Date(incident.incidentStart).toLocaleString()} to {new Date(incident.incidentEnd).toLocaleString()}</dd></div>
      </dl>
      <code>{item.query}</code>
    </details>
  </article>
}

function InvestigationPanel({ incident, onSubmit }: { incident: Incident; onSubmit: WorkspaceProps["onSubmit"] }) {
  return <div className="investigation-panel">
    <div className="form-grid">
      <form onSubmit={onSubmit("hypotheses", form => ({
        summary: form.get("summary"),
        confidence: form.get("confidence"),
        evidenceIds: incident.evidence.slice(0, 1).map(item => item.id),
        alternatives: [String(form.get("alternative") || "Unconfirmed alternative")],
        nextStep: form.get("nextStep"),
      }), "Hypothesis saved.")}>
        <FormTitle title="Record hypothesis" detail="Link the explanation to measured evidence."/>
        <Field label="Hypothesis"><input name="summary" required placeholder="Example: one query runs for each cart item"/></Field>
        <div className="field-row">
          <Field label="Confidence"><select name="confidence" defaultValue="medium"><option>low</option><option>medium</option><option>high</option></select></Field>
          <Field label="Evidence"><input value={incident.evidence[0]?.id ?? "No evidence"} readOnly/></Field>
        </div>
        <Field label="Alternative"><input name="alternative" required placeholder="Example: database saturation"/></Field>
        <Field label="Next step"><input name="nextStep" required placeholder="Example: inspect the product lookup loop"/></Field>
        <button className="secondary-button">Save hypothesis</button>
      </form>
      <form onSubmit={onSubmit("changes", form => ({
        summary: form.get("summary"),
        files: String(form.get("files") || "").split("\n").filter(Boolean),
        tests: String(form.get("tests") || "").split("\n").filter(Boolean),
      }), "Change saved.")}>
        <FormTitle title="Record code change" detail="List the implementation and the tests that passed."/>
        <Field label="Change"><input name="summary" required placeholder="Example: replaced item lookups with a batch query"/></Field>
        <Field label="Files"><textarea name="files" required placeholder="One file per line"/></Field>
        <Field label="Tests"><textarea name="tests" required placeholder="One test command per line"/></Field>
        <button className="secondary-button">Save change</button>
      </form>
    </div>
    {(incident.hypotheses.length > 0 || incident.changes.length > 0) && <div className="saved-records">
      {incident.hypotheses.map(item => <article key={item.id}><span>Hypothesis · {item.confidence}</span><h3>{item.summary}</h3><p>Alternative: {item.alternatives.join(", ")}</p><small>Evidence {item.evidenceIds.join(", ")} · Next {item.nextStep}</small></article>)}
      {incident.changes.map(item => <article key={item.id}><span>Code change</span><h3>{item.summary}</h3><p>{item.files.join(", ") || "No files listed"}</p><small>{item.tests.join(", ") || "No tests listed"}</small></article>)}
    </div>}
    <form className="note-form" onSubmit={onSubmit("notes", form => ({ note: form.get("note"), actor: "Operator" }), "Note saved.")}>
      <Field label="Add operator context"><input name="note" required placeholder="Information the telemetry or agent cannot infer"/></Field>
      <button className="quiet-button">Add note</button>
    </form>
  </div>
}

function RecoveryPanel({ incident, busy, onVerify }: { incident: Incident; busy: boolean; onVerify: () => void }) {
  if (!incident.changes.length) return <EmptyPanel title="No recorded change" text="Record the code change and tests before checking recovery."/>
  return <div className="recovery-panel">
    <div>
      <span className="section-kicker">Original baseline and tolerances</span>
      <h2>{incident.verification ? label(incident.verification.status) : "Ready to verify"}</h2>
      <p>{incident.verification ? `Checked ${new Date(incident.verification.checkedAt).toLocaleString()}.` : "Mica will query fresh telemetry and apply the incident's saved success criteria."}</p>
      <button className="primary-button" onClick={onVerify} disabled={busy}>{busy ? "Checking…" : incident.verification ? "Check again" : "Verify recovery"}</button>
    </div>
    {incident.verification && <div className="verification-grid">{incident.verification.signals.map(signal => <article key={signal.id}><span>{signal.signal}</span><b>{measure(signal.currentValue ?? signal.incidentValue, signal.unit)}</b><small className={signal.classification}>{label(signal.classification)}</small></article>)}</div>}
  </div>
}

function ContextPanel({ service }: { service: Service | null }) {
  if (!service) return <EmptyPanel title="Service context unavailable" text="Select a configured service to inspect its context."/>
  return <div className="context-grid">
    <ContextItem label="Repository" value={service.repository || "Not configured"}/>
    <ContextItem label="Owner" value={service.owners?.join(", ") || "Not configured"}/>
    <ContextItem label="Runtime" value={[service.runtime, service.framework].filter(Boolean).join(" · ") || "Not configured"}/>
    <ContextItem label="Dependencies" value={service.dependencies?.join(", ") || "None configured"}/>
    <ContextItem label="Runbooks" value={service.runbooks?.join(", ") || "Not configured"}/>
    <ContextItem label="Source" value={service.source ? `${service.source.kind} · ${new Date(service.source.refreshedAt).toLocaleString()}` : "Not configured"}/>
  </div>
}

function SharePanel(props: WorkspaceProps) {
  const { incident, destinations, chosen } = props
  return <div className="share-panel">
    <div className="share-controls">
      <FormTitle title="Prepare update" detail="The preview is generated from this incident record."/>
      <div className="field-row">
        <Field label="Update type"><select value={props.updateType} onChange={event => props.onUpdateType(event.target.value)}><option value="incident_opened">Incident opened</option><option value="investigation_update">Investigation update</option><option value="approval_needed">Approval needed</option><option value="mitigation_applied">Mitigation applied</option><option value="recovery_verified">Recovery verified</option><option value="handoff">Handoff</option></select></Field>
        <Field label="Audience"><select value={props.audience} onChange={event => props.onAudience(event.target.value)}><option value="engineering">Engineering</option><option value="stakeholders">Stakeholders</option><option value="public">Public</option></select></Field>
      </div>
      <fieldset className="destination-list">
        <legend>Destinations</legend>
        {destinations.length ? destinations.map(destination => <label key={destination.id}><input type="checkbox" checked={chosen.includes(destination.id)} onChange={() => props.onChosen(chosen.includes(destination.id) ? chosen.filter(id => id !== destination.id) : [...chosen, destination.id])}/>{destination.provider}</label>) : <p>No destination configured. You can still save a local draft.</p>}
      </fieldset>
      {destinations.length > 0 && <button className="text-button" onClick={() => props.onChosen(destinations.map(item => item.id))}>Select all destinations</button>}
      <button className="secondary-button" onClick={props.onPrepare}>Prepare preview</button>
    </div>
    <div className="update-list">
      {incident.updates.length ? incident.updates.slice().reverse().map(update => <article className="update-card" key={update.id}>
        <div><span>{label(update.updateType)}</span><b>{label(update.status)}</b></div>
        <p>{update.body}</p>
        <small>Audience {update.audience} · Redactions {update.redactions.join(", ") || "none"}</small>
        <div className="receipts">{update.destinations.map(receipt => <span key={receipt.destinationId}>{receipt.provider}: {label(receipt.status)}</span>)}</div>
        {(update.status === "pending_approval" || update.status === "partially_delivered") && <div className="approval-row"><Field label="Approver"><input value={props.approver} onChange={event => props.onApprover(event.target.value)} placeholder="Name"/></Field><button className="primary-button" onClick={() => props.onPublish(update)}>{update.status === "partially_delivered" ? "Retry failed" : "Approve and send"}</button></div>}
      </article>) : <EmptyPanel title="No update drafts" text="Prepare a preview when the incident record is ready to share."/>}
    </div>
  </div>
}

function ReviewPanel({ incident, onReview }: { incident: Incident; onReview: WorkspaceProps["onReview"] }) {
  return <div className="review-grid">
    <section>
      <FormTitle title="Action proposals" detail="Review changes to Mica records. No infrastructure action is executed."/>
      {incident.proposals.length ? incident.proposals.map(proposal => <article className="review-card" key={proposal.id}>
        <div><span>{proposal.riskLevel}</span><b>{label(proposal.status)}</b></div>
        <h3>{proposal.summary}</h3>
        <dl><div><dt>Expected effect</dt><dd>{proposal.expectedEffect}</dd></div><div><dt>Risk</dt><dd>{proposal.riskStatement}</dd></div><div><dt>Verify</dt><dd>{proposal.verificationPlan}</dd></div></dl>
        {proposal.status === "pending_review" && <div className="review-actions"><button onClick={() => onReview(proposal, "reviewed")}>Mark reviewed</button><button onClick={() => onReview(proposal, "deferred")}>Defer</button><button onClick={() => onReview(proposal, "rejected")}>Reject</button></div>}
      </article>) : <EmptyPanel title="No action proposals" text="Agent proposals will appear here with risk and verification plans."/>}
    </section>
    <section>
      <FormTitle title="Audit findings" detail="Findings remain separate from incident evidence."/>
      {incident.auditFindings.length ? incident.auditFindings.map(finding => <article className={`review-card finding-${finding.severity}`} key={finding.id}><div><span>{finding.severity}</span><b>{label(finding.status)}</b></div><h3>{finding.category}</h3><p>{finding.assessment}</p><small>Evidence {finding.evidenceRefs.join(", ") || "unknown"} · Verify {finding.verificationMethod}</small><p>Recommendation: {finding.recommendation}</p></article>) : <EmptyPanel title="No audit findings" text="Readiness, security, and release reviews will appear here."/>}
    </section>
  </div>
}

function PanelTab({ active, onClick, label: text, count }: { active: boolean; onClick: () => void; label: string; count?: number }) {
  return <button className={active ? "active" : ""} aria-current={active ? "page" : undefined} onClick={onClick}>{text}{count !== undefined && <span>{count}</span>}</button>
}

function SectionHeader({ title, detail }: { title: string; detail: string }) {
  return <div className="section-header"><h2>{title}</h2><span>{detail}</span></div>
}

function FormTitle({ title, detail }: { title: string; detail: string }) {
  return <div className="form-title"><h2>{title}</h2><p>{detail}</p></div>
}

function Field({ label: text, children }: { label: string; children: ReactNode }) {
  return <label className="field"><span>{text}</span>{children}</label>
}

function ContextItem({ label: text, value }: { label: string; value: string }) {
  return <div className="context-item"><span>{text}</span><b>{value}</b></div>
}

function EmptyPanel({ title, text }: { title: string; text: string }) {
  return <div className="empty-panel"><h3>{title}</h3><p>{text}</p></div>
}
