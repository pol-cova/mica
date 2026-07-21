# Mica demo film

Remotion source and real product captures for the Mica hackathon demo.

## Deliverables

- `out/mica-demo.mp4` — 1920×1080 H.264/AAC, 2:59.5
- `out/mica-demo-thumbnail.png` — YouTube thumbnail frame
- `out/mica-demo-contact-sheet.png` — export review sheet

The product captures were collected from the running local Mica stack through Computer Use. They show the real landing page, Prometheus comparison, evidence charts, agent handoff, recovery gate, setup, and documentation.

## Edit or render

```bash
npm install
npm run dev
npx remotion render MicaDemo out/mica-demo.mp4 --codec=h264 --crf=18 --audio-bitrate=192k
```

The final narration uses the selected ElevenLabs Samantha generation, normalized to broadcast-safe levels and adjusted to 1.05× tempo without changing pitch. Scene boundaries follow its natural pauses. The mix adds restrained transition cues and a quiet ambient bed. Timing and visual composition live in `src/Composition.tsx`.

## Submission fit

The film is under three minutes and its narration explains how Codex and GPT-5.6 accelerated architecture, implementation, testing, documentation, and the video itself. The closing frame contains the explicit GPT-5.6-Sol in Codex disclosure.
