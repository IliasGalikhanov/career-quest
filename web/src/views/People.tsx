import { useEffect, useState } from 'react'
import { Users, Compass, CheckCircle2 } from 'lucide-react'
import { api, problemOf } from '../api'
import { localize } from '../i18n'
import type { HR, Problem } from '../types'
import type { ViewProps } from './shared'

export default function People({ dataset, language, t }: ViewProps) {
  const [department, setDepartment] = useState('')
  const [role, setRole] = useState('')
  const [result, setResult] = useState<HR | null>(null)
  const [problem, setProblem] = useState<Problem | null>(null)
  useEffect(() => {
    let active = true
    const query = new URLSearchParams({ department, role })
    api<HR>(`/datasets/${dataset.id}/hr?${query}`).then(value => { if (active) { setResult(value); setProblem(null) } }).catch(error => { if (active) setProblem(problemOf(error)) })
    return () => { active = false }
  }, [dataset, department, role])
  const name = (id: string) => { const skill = dataset.bundle.skills.find(s => s.skill_id === id); return localize(skill?.names, skill?.name ?? id, language) }
  const title = (id: string) => { const event = dataset.bundle.events.find(e => e.event_id === id); return localize(event?.titles, event?.title ?? id, language) }
  return <><header className="page-heading"><div><p className="eyebrow">{t('hr')}</p><h1>{t('insightsTitle')}</h1><p className="muted">{t('insightsBody')}</p></div></header>
    <div className="filters"><label>{t('department')}<select value={department} onChange={e => { setResult(null); setDepartment(e.target.value) }}><option value="">{t('all')}</option>{[...new Set(dataset.bundle.employees.map(e => e.department))].map(d => <option key={d}>{d}</option>)}</select></label><label>{t('filterRole')}<select value={role} onChange={e => { setResult(null); setRole(e.target.value) }}><option value="">{t('all')}</option>{[...new Set(dataset.bundle.employees.map(e => e.role))].map(r => <option key={r}>{r}</option>)}</select></label></div>
    {problem ? <div className="error-box" role="alert">{t(problem.code)}</div> : !result ? <div className="loading">{t('loading')}</div> : <>
      <div className="stat-strip"><div><Users className="stat-symbol" /><strong>{result.people}</strong><span>{t('people')}</span></div><div><Compass className="stat-symbol" /><strong>{result.without_step.length}</strong><span>{t('noStep')}</span></div><div><CheckCircle2 className="stat-symbol" /><strong>{result.statuses.completed ?? 0}</strong><span>{t('completed')}</span></div></div>
      <div className="hr-grid"><section className="panel"><div className="section-heading"><h2>{t('gapPeople')}</h2></div><div className="bar-list">{result.gaps.filter(g => g.people > 0).map(g => <div key={g.skill_id}><div className="bar-heading"><strong>{name(g.skill_id)}</strong><span>{g.people} / {g.eligible}</span></div><div className="bar-track"><div className="bar-fill" style={{ width: `${g.eligible ? 100 * g.people / g.eligible : 0}%` }} /></div><small className="muted">{t('totalGap')}: {g.total_gap}</small></div>)}{result.gaps.every(g => g.people === 0) && <p className="muted">{t('noGaps')}</p>}</div></section>
      <section className="panel"><div className="section-heading"><h2>{t('noStep')}</h2></div><div className="no-step-list">{result.without_step.map(e => <article key={e.employee_id}><strong>{e.name}</strong><p>{e.reasons.map(reason => t(reason)).join(' · ')}</p></article>)}{result.without_step.length === 0 && <p className="muted">{t('noStepEmpty')}</p>}</div></section></div>
      <section className="panel"><div className="section-heading"><h2>{t('participation')}</h2></div><div className="table-scroll"><table><thead><tr><th>{t('event')}</th>{['completed', 'in_progress', 'dropped', 'declined', 'no_show', 'overdue'].map(status => <th key={status}>{t(status)}</th>)}</tr></thead><tbody>{result.participation.map(p => <tr key={p.event_id}><td>{title(p.event_id)}</td>{['completed', 'in_progress', 'dropped', 'declined', 'no_show', 'overdue'].map(status => <td key={status}>{p.statuses[status] ?? 0}</td>)}</tr>)}</tbody></table></div>{result.participation.length === 0 && <p className="muted">{t('empty')}</p>}</section>
    </>}
  </>
}
