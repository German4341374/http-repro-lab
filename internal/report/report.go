package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/German4341374/http-repro-lab/internal/model"
)

type Data struct {
	Analysis   model.Analysis    `json:"analysis"`
	Comparison *model.Comparison `json:"comparison,omitempty"`
	Generated  map[string]string `json:"generated,omitempty"`
}

func Write(directory string, data Data) error {
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return err
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	files := map[string][]byte{"index.html": []byte(indexHTML), "app.js": []byte(appJS), "styles.css": []byte(stylesCSS), "data.js": append([]byte("window.REPORT_DATA = "), append(raw, []byte(";\n")...)...)}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(directory, name), content, 0o640); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	return nil
}

const indexHTML = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta name="color-scheme" content="dark light"><title>HTTP Repro Lab Report</title><link rel="stylesheet" href="styles.css"></head>
<body><header><div><span class="eyebrow">OFFLINE · PRIVACY REVIEW</span><h1>HTTP Repro Lab</h1><p>Evidence-backed HTTP reproduction report</p></div><button id="theme" type="button">Toggle theme</button></header>
<main><section id="stats" class="stats" aria-label="Summary"></section><nav aria-label="Report sections"><button data-tab="requests" aria-selected="true">Requests</button><button data-tab="privacy">Sensitive values</button><button data-tab="findings">Findings</button><button data-tab="comparison">Comparison</button><button data-tab="code">Generated code</button></nav><label class="search">Filter <input id="search" type="search" placeholder="method, host, path, finding"></label><section id="content" tabindex="-1"></section></main>
<footer>No telemetry. No external assets. Review sanitized artifacts before sharing.</footer><script src="data.js"></script><script src="app.js"></script></body></html>`

const stylesCSS = `:root{font-family:Inter,ui-sans-serif,system-ui,sans-serif;color-scheme:dark;--bg:#071018;--panel:#101d29;--line:#26394a;--text:#eef7ff;--muted:#9eb2c4;--accent:#5eead4;--warn:#fbbf24}*{box-sizing:border-box}body{margin:0;background:radial-gradient(circle at 80% 0,#123454 0,transparent 38%),var(--bg);color:var(--text)}body.light{color-scheme:light;--bg:#f5f8fb;--panel:#fff;--line:#cbd5e1;--text:#102235;--muted:#52677a;--accent:#087f75}header,main,footer{max-width:1180px;margin:auto;padding:24px}header{display:flex;justify-content:space-between;align-items:start;border-bottom:1px solid var(--line)}h1{font-size:clamp(2rem,6vw,4.5rem);margin:.2rem 0;letter-spacing:-.06em}.eyebrow{color:var(--accent);font-weight:800;letter-spacing:.16em;font-size:.75rem}button,input{font:inherit}button{color:var(--text);background:var(--panel);border:1px solid var(--line);border-radius:9px;padding:.65rem .9rem;cursor:pointer}button[aria-selected=true]{background:var(--accent);color:#052620;border-color:transparent}.stats{display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:12px;margin:18px 0}.stat,.card{background:color-mix(in srgb,var(--panel) 93%,transparent);border:1px solid var(--line);border-radius:14px;padding:18px;box-shadow:0 16px 50px #0002}.stat strong{font-size:1.8rem;display:block}.stat span,.muted{color:var(--muted)}nav{display:flex;gap:8px;flex-wrap:wrap;margin:18px 0}.search{display:block;margin:14px 0;color:var(--muted)}input{width:min(440px,100%);display:block;margin-top:6px;background:var(--panel);color:var(--text);border:1px solid var(--line);border-radius:9px;padding:.75rem}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:12px}.method{font-weight:900;color:var(--accent)}code,pre{font-family:ui-monospace,SFMono-Regular,Consolas,monospace}pre{white-space:pre-wrap;overflow:auto;background:#03080d;color:#dff9f5;border-radius:10px;padding:14px;max-height:440px}.badge{display:inline-block;background:#25384a;border-radius:999px;padding:.25rem .55rem;margin:.15rem;font-size:.8rem}.warning{color:var(--warn)}details{margin-top:10px}footer{color:var(--muted);font-size:.85rem}@media(max-width:600px){header{display:block}header button{margin-top:12px}}`

const appJS = `(() => {
  'use strict';
  const data = window.REPORT_DATA || { analysis: { requests: [], findings: [], sensitiveValues: [] } };
  let tab = 'requests';
  let query = '';
  const content = document.querySelector('#content');
  const text = (tag, value, className) => { const node = document.createElement(tag); node.textContent = String(value ?? ''); if (className) node.className = className; return node; };
  const card = () => { const node = document.createElement('article'); node.className = 'card'; return node; };
  const matches = (value) => JSON.stringify(value).toLowerCase().includes(query);
  function stats() { const a=data.analysis; const items=[[a.requests.length,'requests'],[a.sensitiveValues.length,'sensitive values'],[a.findings.length,'findings'],[data.comparison?.differences?.length||0,'differences']]; const root=document.querySelector('#stats'); root.replaceChildren(...items.map(([n,l])=>{const d=text('div','', 'stat');d.append(text('strong',n),text('span',l));return d;})); }
  function renderRequests(){ const root=text('div','', 'grid'); data.analysis.requests.filter(matches).forEach((r,i)=>{const c=card();c.append(text('span',r.method,'method'),text('h2',r.url.host+r.url.path),text('p',r.url.scheme+'://'+r.url.host,'muted'),text('span','Request '+(i+1),'badge'));const details=document.createElement('details');details.append(text('summary','Normalized RequestSpec'),text('pre',JSON.stringify(r,null,2)));c.append(details);root.append(c);});return root; }
  function renderPrivacy(){ const root=text('div','', 'grid'); data.analysis.sensitiveValues.filter(matches).forEach(v=>{const c=card();c.append(text('span',v.confidence,'badge'),text('h2',v.type),text('p',v.location),text('code',v.preview),text('p',v.suggestedAction,'muted'));root.append(c);});return root; }
  function renderFindings(){ const root=text('div','', 'grid'); data.analysis.findings.filter(matches).forEach(f=>{const c=card();c.append(text('span',f.severity,'badge'),text('h2',f.title),text('p',f.summary),text('p','Confidence: '+f.confidence,'muted'));(f.evidence||[]).forEach(e=>c.append(text('p',e.source+' · '+e.field+' · '+e.value)));root.append(c);});return root; }
  function renderComparison(){ const root=text('div','', 'grid'); (data.comparison?.differences||[]).filter(matches).forEach(d=>{const c=card();c.append(text('h2',d.field),text('p',d.evidence),text('pre',JSON.stringify({environmentA:d.environmentA,environmentB:d.environmentB},null,2)),text('p',d.possibleInterpretation),text('p','Verify: '+d.suggestedVerification,'muted'));root.append(c);});return root; }
  function renderCode(){ const root=text('div','', 'grid'); Object.entries(data.generated||{}).filter(matches).forEach(([language,source])=>{const c=card();c.append(text('h2',language),text('pre',source));root.append(c);});return root; }
  function render(){ const views={requests:renderRequests,privacy:renderPrivacy,findings:renderFindings,comparison:renderComparison,code:renderCode};content.replaceChildren(views[tab]()); }
  document.querySelectorAll('[data-tab]').forEach(button=>button.addEventListener('click',()=>{tab=button.dataset.tab;document.querySelectorAll('[data-tab]').forEach(b=>b.setAttribute('aria-selected',String(b===button)));render();content.focus();}));
  document.querySelector('#search').addEventListener('input',event=>{query=event.target.value.toLowerCase();render();}); document.querySelector('#theme').addEventListener('click',()=>document.body.classList.toggle('light'));
  stats(); render();
})();`
