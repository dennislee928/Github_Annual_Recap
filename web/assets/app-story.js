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
  if (count <= 0) return 0;
  if (count <= 1) return 1;
  if (count <= 3) return 2;
  if (count <= 6) return 3;
  return 4;
}

function byDesc(a,b){ return (b[1]||0) - (a[1]||0); }

function getLanguageIcon(lang) {
  const iconMap = {
    'Python': 'language-python',
    'JavaScript': 'language-js',
    'TypeScript': 'language-ts',
    'Go': 'language-go',
    'CSS': 'language-css',
    'HTML': 'language-html',
  };
  return iconMap[lang] || null;
}

function calculateAchievements(recap) {
  const achievements = [];
  
  // Streak achievements
  const streak = recap.calendar?.longest_streak || 0;
  if (streak >= 365) achievements.push({ type: 'streak', label: 'Year Streak', value: '365+ days', icon: '🔥' });
  else if (streak >= 100) achievements.push({ type: 'streak', label: 'Century Streak', value: '100+ days', icon: '💯' });
  else if (streak >= 30) achievements.push({ type: 'streak', label: 'Monthly Streak', value: '30+ days', icon: '📅' });
  else if (streak >= 7) achievements.push({ type: 'streak', label: 'Weekly Streak', value: '7+ days', icon: '⭐' });
  
  // PR achievements
  const prs = recap.totals?.pull_requests || 0;
  if (prs >= 2000) achievements.push({ type: 'pr', label: 'PR Master', value: '2000+ PRs', icon: '🚀' });
  else if (prs >= 1000) achievements.push({ type: 'pr', label: 'PR Expert', value: '1000+ PRs', icon: '💪' });
  else if (prs >= 500) achievements.push({ type: 'pr', label: 'PR Pro', value: '500+ PRs', icon: '🎯' });
  else if (prs >= 100) achievements.push({ type: 'pr', label: 'PR Contributor', value: '100+ PRs', icon: '✨' });
  
  // Contribution achievements
  const total = recap.totals?.overall || 0;
  if (total >= 10000) achievements.push({ type: 'contrib', label: '10K Club', value: '10,000+ contributions', icon: '🏆' });
  else if (total >= 5000) achievements.push({ type: 'contrib', label: '5K Club', value: '5,000+ contributions', icon: '🥇' });
  else if (total >= 1000) achievements.push({ type: 'contrib', label: '1K Club', value: '1,000+ contributions', icon: '🥈' });
  
  // Language diversity
  const langCount = Object.keys(recap.languages?.weighted_bytes || {}).length;
  if (langCount >= 10) achievements.push({ type: 'lang', label: 'Polyglot', value: '10+ languages', icon: '🌍' });
  else if (langCount >= 5) achievements.push({ type: 'lang', label: 'Multilingual', value: '5+ languages', icon: '🗣️' });
  
  // Commit achievements
  const commits = recap.totals?.commits || 0;
  if (commits >= 5000) achievements.push({ type: 'commit', label: 'Commit Master', value: '5000+ commits', icon: '💻' });
  else if (commits >= 1000) achievements.push({ type: 'commit', label: 'Commit Expert', value: '1000+ commits', icon: '⌨️' });
  
  return achievements;
}

function calculateAdditionalStats(recap) {
  const stats = {};
  
  // Total repos contributed to
  const repoSet = new Set();
  (recap.top_repos || []).forEach(r => repoSet.add(r.repo));
  stats.totalRepos = repoSet.size;
  
  // Average contributions per day
  const daysInYear = 365;
  const totalContribs = recap.totals?.overall || 0;
  stats.avgPerDay = (totalContribs / daysInYear).toFixed(1);
  
  // Most active month
  const monthCounts = {};
  (recap.calendar?.days || []).forEach(d => {
    if (d.date) {
      const month = d.date.substring(0, 7); // YYYY-MM
      monthCounts[month] = (monthCounts[month] || 0) + (d.count || 0);
    }
  });
  const sortedMonths = Object.entries(monthCounts).sort((a, b) => b[1] - a[1]);
  if (sortedMonths.length > 0) {
    const [month, count] = sortedMonths[0];
    const monthNames = ['', 'January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];
    const monthNum = parseInt(month.split('-')[1]);
    stats.mostActiveMonth = `${monthNames[monthNum]} ${month.split('-')[0]} (${fmt(count)} contributions)`;
  } else {
    stats.mostActiveMonth = "—";
  }
  
  // Contribution distribution
  const commits = recap.totals?.commits || 0;
  const prs = recap.totals?.pull_requests || 0;
  const issues = recap.totals?.issues || 0;
  const total = commits + prs + issues;
  stats.distribution = [
    { label: 'Commits', value: commits, percent: total > 0 ? (commits / total * 100).toFixed(1) : 0 },
    { label: 'Pull Requests', value: prs, percent: total > 0 ? (prs / total * 100).toFixed(1) : 0 },
    { label: 'Issues', value: issues, percent: total > 0 ? (issues / total * 100).toFixed(1) : 0 },
  ];
  
  return stats;
}

function renderBadges(containerId, achievements) {
  const container = document.getElementById(containerId);
  if (!container) return;
  
  container.innerHTML = "";
  achievements.forEach(ach => {
    const badge = document.createElement("div");
    badge.className = "badge-item";
    badge.innerHTML = `
      <span class="badge-icon">${ach.icon}</span>
      <span>${ach.label}: ${ach.value}</span>
    `;
    container.appendChild(badge);
  });
}

async function main(){
  const dataFile = qs("data") || "recap_2025.json";
  const res = await fetch(`./${dataFile}`);
  if (!res.ok) throw new Error(`Failed to load ${dataFile}: ${res.status}`);
  const recap = await res.json();

  // Calculate achievements
  const achievements = calculateAchievements(recap);
  
  // Calculate additional stats
  const additionalStats = calculateAdditionalStats(recap);

  // Card 01: You shipped
  document.getElementById("t_commits").textContent = fmt(recap.totals.commits);
  document.getElementById("t_prs").textContent = fmt(recap.totals.pull_requests);
  document.getElementById("t_issues").textContent = fmt(recap.totals.issues);
  document.getElementById("t_reviews").textContent = fmt(recap.totals.reviews);
  document.getElementById("t_overall").textContent = fmt(recap.totals.overall);
  renderBadges("badges-01", achievements.filter(a => a.type === 'contrib' || a.type === 'commit' || a.type === 'pr'));

  // Card 02: Calendar heatmap
  document.getElementById("streak").textContent = fmt(recap.calendar.longest_streak);
  document.getElementById("best_day").textContent = `${recap.calendar.most_productive_day.date} (${fmt(recap.calendar.most_productive_day.count)})`;
  document.getElementById("best_week").textContent = `${recap.calendar.most_productive_iso_week.iso_week} (${fmt(recap.calendar.most_productive_iso_week.count)})`;
  
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
  renderBadges("badges-02", achievements.filter(a => a.type === 'streak'));

  // Card 03: Top repos
  const repoList = document.getElementById("top_repos");
  repoList.innerHTML = "";
  (recap.top_repos || []).slice(0, 10).forEach((r) => {
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

  // Card 04: PR stats
  document.getElementById("pr_opened").textContent = fmt(recap.pr_stats.opened);
  document.getElementById("pr_merged").textContent = fmt(recap.pr_stats.merged);
  document.getElementById("pr_merge_rate").textContent = (recap.pr_stats.merge_rate != null) ? (recap.pr_stats.merge_rate*100).toFixed(1) + "%" : "—";
  document.getElementById("pr_avg_merge").textContent = hoursToHuman(recap.pr_stats.avg_time_to_merge_hours);
  
  const big = recap.pr_stats.biggest_pr;
  document.getElementById("big_pr").textContent = big ? `${big.repo}#${big.number} (+${fmt(big.additions)}/-${fmt(big.deletions)})` : "—";
  renderBadges("badges-04", achievements.filter(a => a.type === 'pr'));

  // Card 05: Issues + reviews
  document.getElementById("iss_opened").textContent = fmt(recap.issue_stats.opened);
  document.getElementById("iss_closed").textContent = fmt(recap.issue_stats.closed);
  const t2 = document.getElementById("t_reviews2");
  if (t2) t2.textContent = fmt(recap.totals.reviews);
  
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

  // Card 06: Languages
  const langs = recap.languages?.weighted_bytes || {};
  const items = Object.entries(langs).map(([k,v])=>[k, Number(v)]).sort(byDesc).slice(0, 10);
  const langList = document.getElementById("langs");
  langList.innerHTML = "";
  const total = items.reduce((a, [,v])=>a+v, 0) || 1;
  items.forEach(([k,v])=>{
    const item = document.createElement("div");
    item.className = "story-lang-item";
    const iconName = getLanguageIcon(k);
    if (iconName) {
      const icon = document.createElement("img");
      icon.src = `./assets/icons/${iconName}.svg`;
      icon.className = "story-lang-icon";
      icon.alt = k;
      item.appendChild(icon);
    }
    const name = document.createElement("div");
    name.className = "story-lang-name";
    name.textContent = k;
    const meta = document.createElement("div");
    meta.className = "story-lang-meta";
    meta.textContent = `${(v/total*100).toFixed(1)}% (${fmt(v)})`;
    item.appendChild(name);
    item.appendChild(meta);
    langList.appendChild(item);
  });
  renderBadges("badges-06", achievements.filter(a => a.type === 'lang'));

  // Card 07: Stars + forks
  const g = recap.growth;
  if (g) {
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

  // Card 08: Additional stats - Repos
  document.getElementById("total_repos").textContent = fmt(additionalStats.totalRepos);
  document.getElementById("avg_per_day").textContent = additionalStats.avgPerDay;

  // Card 09: Additional stats - Activity timeline
  document.getElementById("most_active_month").textContent = additionalStats.mostActiveMonth;
  const distContainer = document.getElementById("contribution_dist");
  distContainer.innerHTML = "";
  additionalStats.distribution.forEach(dist => {
    const item = document.createElement("div");
    item.className = "dist-item";
    const label = document.createElement("div");
    label.className = "dist-label";
    label.textContent = dist.label;
    const bar = document.createElement("div");
    bar.className = "dist-bar";
    const fill = document.createElement("div");
    fill.className = "dist-bar-fill";
    fill.style.width = `${dist.percent}%`;
    bar.appendChild(fill);
    const value = document.createElement("div");
    value.className = "dist-value";
    value.textContent = `${dist.percent}% (${fmt(dist.value)})`;
    item.appendChild(label);
    item.appendChild(bar);
    item.appendChild(value);
    distContainer.appendChild(item);
  });

  // Card 10: Achievements
  const achievementsContainer = document.getElementById("achievements");
  achievementsContainer.innerHTML = "";
  achievements.slice(0, 6).forEach(ach => {
    const item = document.createElement("div");
    item.className = "achievement-item";
    item.innerHTML = `
      <div class="achievement-icon">${ach.icon}</div>
      <div class="achievement-label">${ach.label}</div>
      <div class="achievement-value">${ach.value}</div>
    `;
    achievementsContainer.appendChild(item);
  });
}

main().catch((e)=>{
  console.error(e);
  const errEl = document.getElementById("err");
  if (errEl) errEl.textContent = String(e);
});
