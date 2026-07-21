import {Audio} from '@remotion/media';
import React from 'react';
import {
  AbsoluteFill,
  Composition,
  Easing,
  Img,
  interpolate,
  Sequence,
  spring,
  staticFile,
  useCurrentFrame,
  useVideoConfig,
} from 'remotion';

const FPS = 30;
const INK = '#0a1020';
const PAPER = '#f7f9fc';

const scenes = [
  {id: 'hook', frames: 586, caption: 'What changed? What was healthy? Did the fix work?'},
  {id: 'system', frames: 516, caption: 'Prometheus → Go daemon → SQLite → Workspace + MCP'},
  {id: 'detect', frames: 740, caption: 'A healthy baseline compared with the incident window.'},
  {id: 'evidence', frames: 647, caption: 'Every claim stays inspectable.'},
  {id: 'agent', frames: 702, caption: 'The handoff carries service, incident, evidence, and workflow IDs.'},
  {id: 'control', frames: 623, caption: 'Agents investigate. Humans retain execution control.'},
  {id: 'build', frames: 781, caption: 'Codex accelerated architecture, implementation, tests, skills, and docs.'},
  {id: 'close', frames: 787, caption: 'From production regression to verified recovery.'},
] as const;

const totalFrames = scenes.reduce((sum, scene) => sum + scene.frames, 0);

const fade = (frame: number, duration: number) =>
  interpolate(frame, [0, 16, duration - 18, duration], [0, 1, 1, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

const reveal = (frame: number, start: number, duration = 26) =>
  interpolate(frame, [start, start + duration], [0, 1], {
    easing: Easing.bezier(0.16, 1, 0.3, 1),
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

const Grid: React.FC<{dark?: boolean}> = ({dark = false}) => (
  <AbsoluteFill
    style={{
      backgroundColor: dark ? INK : PAPER,
      backgroundImage: dark
        ? 'linear-gradient(rgba(255,255,255,.055) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,.055) 1px, transparent 1px)'
        : 'linear-gradient(rgba(10,16,32,.065) 1px, transparent 1px), linear-gradient(90deg, rgba(10,16,32,.065) 1px, transparent 1px)',
      backgroundSize: '96px 96px',
    }}
  />
);

const Chrome: React.FC<{src: string; zoom?: number; x?: number; y?: number}> = ({
  src,
  zoom = 1,
  x = 0,
  y = 0,
}) => {
  const frame = useCurrentFrame();
  const drift = interpolate(frame, [0, 500], [0, 18], {extrapolateRight: 'clamp'});
  return (
    <div className="browser" style={{transform: `translate(${x}px, ${y - drift}px) scale(${zoom})`}}>
      <div className="browser-bar">
        <span /><span /><span />
        <div>127.0.0.1:8787</div>
        <b><i /> live product</b>
      </div>
      <div className="browser-image"><Img src={staticFile(src)} /></div>
    </div>
  );
};

const Kicker: React.FC<{children: React.ReactNode; dark?: boolean}> = ({children, dark}) => (
  <div className={`kicker ${dark ? 'kicker-dark' : ''}`}><i />{children}</div>
);

const ChapterFooter: React.FC<{text: string; current: number}> = ({text, current}) => (
  <div className="caption-wrap">
    <div className="caption"><span>0{current + 1}</span>{text}</div>
    <div className="timeline"><div style={{width: `${((current + 1) / scenes.length) * 100}%`}} /></div>
  </div>
);

const SceneWipe: React.FC<{dark: boolean}> = ({dark}) => {
  const frame = useCurrentFrame();
  const travel = interpolate(frame, [0, 24], [-18, 112], {
    easing: Easing.bezier(0.22, 1, 0.36, 1),
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const opacity = interpolate(frame, [0, 8, 20, 28], [0, 0.7, 0.35, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  return <div style={{position:'absolute',inset:0,zIndex:18,pointerEvents:'none',overflow:'hidden',opacity}}>
    <div style={{position:'absolute',left:`${travel}%`,top:0,bottom:0,width:260,transform:'skewX(-10deg)',background:dark?'linear-gradient(90deg,transparent,#2367ff,rgba(183,243,74,.7),transparent)':'linear-gradient(90deg,transparent,#2367ff,transparent)',filter:'blur(10px)'}} />
  </div>;
};

const Hook: React.FC = () => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const enter = spring({frame, fps, config: {damping: 18, stiffness: 105}});
  return <AbsoluteFill className="scene dark-scene">
    <Grid dark />
    <div className="orb orb-a" /><div className="orb orb-b" />
    <div className="hook-copy" style={{transform: `translateY(${(1 - enter) * 50}px)`, opacity: enter}}>
      <Kicker dark>PRODUCTION CONTEXT FOR CODING AGENTS</Kicker>
      <h1>Code is only half<br />the <em>incident.</em></h1>
      <p>Connect live evidence, agent work, and recovery in one local record.</p>
    </div>
    <div className="signal-stack">
      {['repository', 'telemetry', 'incident memory', 'verification'].map((x, i) => (
        <div key={x} style={{opacity: interpolate(frame, [25 + i * 9, 42 + i * 9], [0, 1], {extrapolateLeft:'clamp',extrapolateRight:'clamp'}), transform:`translateX(${interpolate(frame,[25+i*9,42+i*9],[70,0],{extrapolateLeft:'clamp',extrapolateRight:'clamp'})}px)`}}>
          <span>0{i + 1}</span>{x}<b>{i === 0 ? 'known' : 'missing'}</b>
        </div>
      ))}
    </div>
  </AbsoluteFill>;
};

const System: React.FC = () => {
  const frame = useCurrentFrame();
  const pop = (delay: number) => spring({frame: frame - delay, fps: FPS, config:{damping:17, stiffness:120}});
  return <AbsoluteFill className="scene light-scene">
    <Grid />
    <div className="section-title"><Kicker>ONE LOCAL SYSTEM</Kicker><h2>Humans and agents<br />share the same incident.</h2></div>
    <div className="architecture">
      <div className="node side" style={{transform:`scale(${pop(16)})`}}><small>INPUT</small><strong>Prometheus</strong><span>metrics + windows</span></div>
      <div className="connector"><i style={{width:`${interpolate(frame,[28,55],[0,100],{extrapolateLeft:'clamp',extrapolateRight:'clamp'})}%`}} /></div>
      <div className="node core" style={{transform:`scale(${pop(28)})`}}><div className="mica-mark">m.</div><small>LOCAL DAEMON</small><strong>Mica</strong><span>Go + SQLite</span></div>
      <div className="branch" style={{opacity:pop(43)}}><div /><div /></div>
      <div className="outputs">
        <div className="node" style={{transform:`scale(${pop(50)})`}}><small>HUMAN</small><strong>Workspace</strong><span>React evidence UI</span></div>
        <div className="node blue" style={{transform:`scale(${pop(59)})`}}><small>AGENT</small><strong>Codex</strong><span>typed MCP tools</span></div>
      </div>
    </div>
  </AbsoluteFill>;
};

const Detect: React.FC = () => {
  const frame = useCurrentFrame();
  const focus = interpolate(frame, [100, 420], [0.94, 1.04], {extrapolateRight:'clamp'});
  return <AbsoluteFill className="scene dark-scene clip">
    <Grid dark />
    <div className="float-title"><Kicker dark>REAL TELEMETRY</Kicker><h2>2 regressions detected</h2><p>Baseline → incident</p></div>
    <Chrome src="captures/workspace.jpg" zoom={focus} x={330} y={310} />
    <div className="metric-callouts">
      {[['P95 LATENCY','47.5','142.8 ms','+201%'],['DB QUERIES / REQUEST','2.0','4.2','+112%']].map((metric,i)=>{
        const progress = reveal(frame, i === 0 ? 330 : 485, 32);
        return <div key={metric[0]} style={{opacity:progress,transform:`translateX(${(1-progress)*-36}px)`}}><small>{metric[0]}</small><strong>{metric[1]} <i>→</i> {metric[2]}</strong><span>{metric[3]}</span></div>;
      })}
    </div>
  </AbsoluteFill>;
};

const Evidence: React.FC = () => {
  const frame = useCurrentFrame();
  const zoom = interpolate(frame,[0,440],[1.1,1.32],{extrapolateRight:'clamp'});
  return <AbsoluteFill className="scene light-scene clip">
    <Grid />
    <div className="top-copy"><Kicker>EVIDENCE, NOT AN ALERT</Kicker><h2>Every claim is inspectable.</h2></div>
    <Chrome src="captures/workspace.jpg" zoom={zoom} x={90} y={250} />
    <div className="evidence-rail">
      {['ev_1 · area trend','ev_2 · line trend','ev_3 · window bars'].map((x,i)=>{const progress=reveal(frame,90+i*105,28);return <div key={x} style={{opacity:progress,transform:`translateX(${(1-progress)*42}px)`}}><span>0{i+1}</span>{x}</div>;})}
      <p style={{opacity:reveal(frame,455,30),transform:`translateY(${(1-reveal(frame,455,30))*24}px)`}}><b>0 baseline?</b><br />No invented ratio.</p>
    </div>
  </AbsoluteFill>;
};

const Agent: React.FC = () => {
  const frame = useCurrentFrame();
  return <AbsoluteFill className="scene dark-scene clip">
    <Grid dark />
    <div className="float-title"><Kicker dark>SHARED WITH CODEX</Kicker><h2>Continue the same incident.</h2><p>Not a second investigation.</p></div>
    <Chrome src="captures/agent.jpg" zoom={0.95} x={365} y={300} />
    <div className="tool-cloud">
      {['get_service_context','inspect_service','record_hypothesis','record_change','verify_recovery'].map((x,i)=>{const progress=reveal(frame,320+i*45,24);return <div key={x} style={{transform:`translateY(${(1-progress)*30}px)`,opacity:progress}}>{x}</div>;})}
    </div>
  </AbsoluteFill>;
};

const Control: React.FC = () => {
  const frame = useCurrentFrame();
  return <AbsoluteFill className="scene light-scene">
    <Grid />
    <div className="section-title compact"><Kicker>HUMAN CONTROL</Kicker><h2>Useful tools.<br />Deliberate limits.</h2></div>
    <div className="boundary">
      <div className="allowed"><small>AGENT CAN</small>{['Read telemetry','Record hypotheses','Prepare actions','Verify recovery'].map((x,i)=><div key={x} style={{opacity:reveal(frame,155+i*42,24)}}><b>✓</b>{x}</div>)}</div>
      <div className="blocked"><small>HUMAN DECIDES</small>{['Deploy','Restart','Rollback','Publish externally'].map((x,i)=><div key={x} style={{opacity:reveal(frame,330+i*36,24)}}><b>—</b>{x}</div>)}</div>
    </div>
    <div className="gate" style={{opacity:reveal(frame,515,28),transform:`translateY(${(1-reveal(frame,515,28))*30}px)`}}><span>RECOVERY GATE</span><strong>No recorded change</strong><p>Verification stays unavailable until the implementation and tests are attached.</p></div>
  </AbsoluteFill>;
};

const Build: React.FC = () => {
  const frame = useCurrentFrame();
  return <AbsoluteFill className="scene dark-scene clip">
    <Grid dark />
    <div className="top-copy dark"><Kicker dark>BUILT WITH GPT-5.6</Kicker><h2>From architecture to a working tool.</h2></div>
    <Chrome src="captures/docs.jpg" zoom={1.02} x={610} y={250} />
    <div className="build-list">
      {['Go daemon','React workspace','MCP tool surface','Agent skills','Docker demo','Tests + docs','Remotion film'].map((x,i)=>{const progress=reveal(frame,95+i*52,26);return <div key={x} style={{opacity:progress,transform:`translateX(${(1-progress)*-30}px)`}}><span>{String(i+1).padStart(2,'0')}</span>{x}</div>;})}
    </div>
    <div className="decision" style={{opacity:reveal(frame,535,30),transform:`translateY(${(1-reveal(frame,535,30))*20}px)`}}>KEY DECISION <b>Local daemon</b><span>Credentials and telemetry stay near the operator.</span></div>
  </AbsoluteFill>;
};

const Close: React.FC = () => {
  const frame = useCurrentFrame();
  const {fps} = useVideoConfig();
  const rise = spring({frame, fps, config:{damping:20,stiffness:90}});
  return <AbsoluteFill className="scene close-scene">
    <Grid dark />
    <div className="close-copy" style={{opacity:rise,transform:`translateY(${(1-rise)*45}px)`}}>
      <div className="logo">mica<span>.</span></div>
      <h2>From production regression<br />to <em>verified recovery.</em></h2>
      <p>Evidence for humans. Context for agents.</p>
      <div className="disclosure" style={{opacity:reveal(frame,560,34),transform:`translateY(${(1-reveal(frame,560,34))*18}px)`}}>BUILT · NARRATED · RECORDED <b>GPT-5.6-SOL IN CODEX</b></div>
    </div>
    <div className="closing-flow"><span>Detect</span><i /><span>Investigate</span><i /><span>Change</span><i /><span>Verify</span></div>
  </AbsoluteFill>;
};

const components = [Hook, System, Detect, Evidence, Agent, Control, Build, Close];

const SceneLayer: React.FC<{index: number}> = ({index}) => {
  const frame = useCurrentFrame();
  const scene = scenes[index];
  const Scene = components[index];
  return <>
    <Scene />
    {index > 0 ? <Audio src={staticFile('audio/whoosh.wav')} volume={0.08} /> : null}
    {index === 4 || index === 5 ? <Audio src={staticFile('audio/switch.wav')} volume={0.12} /> : null}
    {index === 7 ? <Audio src={staticFile('audio/ding.wav')} volume={0.1} /> : null}
    <SceneWipe dark={index === 0 || index === 2 || index === 4 || index === 6 || index === 7} />
    <div style={{opacity:fade(frame, scene.frames)}}><ChapterFooter text={scene.caption} current={index} /></div>
  </>;
};

const Film: React.FC = () => {
  let start = 0;
  return <AbsoluteFill style={{backgroundColor:INK}}>
    <Audio src={staticFile('audio/elevenlabs-master.mp3')} />
    <Audio src={staticFile('audio/ambient-bed.mp3')} volume={0.035} />
    {scenes.map((scene,index)=>{
      const from = start;
      start += scene.frames;
      return <Sequence key={scene.id} from={from} durationInFrames={scene.frames} premountFor={FPS}>
        <SceneLayer index={index} />
      </Sequence>;
    })}
  </AbsoluteFill>;
};

export const MyComposition: React.FC = () => <Composition id="MicaDemo" component={Film} durationInFrames={totalFrames} fps={FPS} width={1920} height={1080} />;
