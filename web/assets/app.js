function qs(name) {
  const u = new URL(window.location.href);
  return u.searchParams.get(name);
}

function fmt(n) {
  return new Intl.NumberFormat("en-US").format(n ?? 0);
}

function hoursToHuman(h) {
  if (!isFinite(h) || h <= 0) return "—";
  if (h < 48) return `${h.toFixed(1)}h`;
  const d = h / 24;
  return `${d.toFixed(1)}d`;
}

function level(count) {
  // 0..4
  if (count <= 0) return 0;
  if (count <= 1) return 1;
  if (count <= 3) return 2;
  if (count <= 6) return 3;
  return 4;
}

function byDesc(a,b){ return (b[1]||0) - (a[1]||0); }

async function main(){
  const dataFile = qs("data") || "recap_2025.json";
  const res = await fetch(`./${dataFile}`);
  if (!res.ok) throw new Error(`Failed to load ${dataFile}: ${res.status}`);
  const recap = await res.json();

  document.getElementById("user").textContent = recap.meta.user;
  document.getElementById("year").textContent = recap.meta.year;
  document.getElementById("gen").textContent = new Date(recap.meta.generated_at).toISOString().replace("T"," ").slice(0,19) + " UTC";

  // Totals card
  document.getElementById("t_commits").textContent = fmt(recap.totals.commits);
  document.getElementById("t_prs").textContent = fmt(recap.totals.pull_requests);
  document.getElementById("t_issues").textContent = fmt(recap.totals.issues);
  document.getElementById("t_reviews").textContent = fmt(recap.totals.reviews);
  const t2 = document.getElementById("t_reviews2");
  if (t2) t2.textContent = fmt(recap.totals.reviews);
  document.getElementById("t_overall").textContent = fmt(recap.totals.overall);

  // Calendar / streak
  document.getElementById("streak").textContent = fmt(recap.calendar.longest_streak);
  document.getElementById("best_day").textContent = `${recap.calendar.most_productive_day.date} (${fmt(recap.calendar.most_productive_day.count)})`;
  document.getElementById("best_week").textContent = `${recap.calendar.most_productive_iso_week.iso_week} (${fmt(recap.calendar.most_productive_iso_week.count)})`;

  // Heatmap (approx. 53 weeks x 7 days; data already flattened; we bucket by ISO week index)
  // Simpler: render 371 cells in order; pad to 53*7=371
  const heat = document.getElementById("heatmap");
  heat.innerHTML = "";
  const days = recap.calendar.days || [];
  const padded = days.slice(0, 371);
  while (padded.length < 371) padded.push({date:"", count:0});
  for (const d of padded) {
    const div = document.createElement("div");
    div.className = "cell";
    div.dataset.l = String(level(d.count||0));
    div.title = d.date ? `${d.date}: ${d.count}` : "";
    heat.appendChild(div);
  }

  // Top repos
  const repoList = document.getElementById("top_repos");
  repoList.innerHTML = "";
  (recap.top_repos || []).forEach((r) => {
    const row = document.createElement("div");
    row.className = "row";
    const left = document.createElement("div");
    left.className = "name";
    left.textContent = r.repo + (r.is_private ? " (private)" : "");
    const right = document.createElement("div");
    right.className = "meta";
    right.textContent = `Activity ${fmt(r.total_activity)} | C ${fmt(r.commit_count)} PR ${fmt(r.pr_count)} I ${fmt(r.issue_count)} R ${fmt(r.review_count)}`;
    row.appendChild(left);
    row.appendChild(right);
    repoList.appendChild(row);
  });

  // PR stats
  document.getElementById("pr_opened").textContent = fmt(recap.pr_stats.opened);
  document.getElementById("pr_merged").textContent = fmt(recap.pr_stats.merged);
  document.getElementById("pr_merge_rate").textContent = (recap.pr_stats.merge_rate != null) ? (recap.pr_stats.merge_rate*100).toFixed(1) + "%" : "—";
  document.getElementById("pr_avg_merge").textContent = hoursToHuman(recap.pr_stats.avg_time_to_merge_hours);

  const big = recap.pr_stats.biggest_pr;
  document.getElementById("big_pr").textContent = big ? `${big.repo}#${big.number} (+${big.additions}/-${big.deletions})` : "—";

  // Issue stats
  document.getElementById("iss_opened").textContent = fmt(recap.issue_stats.opened);
  document.getElementById("iss_closed").textContent = fmt(recap.issue_stats.closed);

  // Languages
  const langs = recap.languages?.weighted_bytes || {};
  const items = Object.entries(langs).map(([k,v])=>[k, Number(v)]).sort(byDesc).slice(0, 10);
  const langList = document.getElementById("langs");
  langList.innerHTML = "";
  const total = items.reduce((a, [,v])=>a+v, 0) || 1;
  items.forEach(([k,v])=>{
    const row = document.createElement("div");
    row.className = "row";
    const left = document.createElement("div");
    left.className = "name";
    left.textContent = k;
    const right = document.createElement("div");
    right.className = "meta";
    right.textContent = `${(v/total*100).toFixed(1)}% (${fmt(v)})`;
    row.appendChild(left);
    row.appendChild(right);
    langList.appendChild(row);
  });


// Reviews by repo (top)
const reviewRepos = recap.reviews?.by_repo || [];
const rr = document.getElementById("review_repos");
if (rr) {
  rr.innerHTML = "";
  reviewRepos.slice(0, 8).forEach((r)=>{
    const row = document.createElement("div");
    row.className = "row";
    const left = document.createElement("div");
    left.className = "name";
    left.textContent = r.repo + (r.is_private ? " (private)" : "");
    const right = document.createElement("div");
    right.className = "meta";
    right.textContent = `${fmt(r.count)} reviews`;
    row.appendChild(left);
    row.appendChild(right);
    rr.appendChild(row);
  });
}

  // Growth
  const g = recap.growth;
  const growthBox = document.getElementById("growth");
  if (!g) {
    growthBox.innerHTML = `<div class="small">Growth metrics were skipped or unavailable. Re-run without <code>--skip-growth</code>.</div>`;
  } else {
    document.getElementById("stars_gained").textContent = fmt(g.total_stars_gained);
    document.getElementById("forks_gained").textContent = fmt(g.total_forks_gained);
    document.getElementById("stars_now").textContent = fmt(g.total_stars_now);
    document.getElementById("forks_now").textContent = fmt(g.total_forks_now);

    const top = (g.repos || []).slice(0, 8);
    const list = document.getElementById("growth_repos");
    list.innerHTML = "";
    top.forEach((r)=>{
      const row = document.createElement("div");
      row.className = "row";
      const left = document.createElement("div");
      left.className = "name";
      left.textContent = r.repo + (r.is_private ? " (private)" : "");
      const right = document.createElement("div");
      right.className = "meta";
      right.textContent = `+★${fmt(r.stars_gained_in_year)} +⑂${fmt(r.forks_gained_in_year)} | now ★${fmt(r.stars_now)} ⑂${fmt(r.forks_now)}`;
      row.appendChild(left);
      row.appendChild(right);
      list.appendChild(row);
    });
  }
}

main().catch((e)=>{
  console.error(e);
  document.getElementById("err").textContent = String(e);
});
