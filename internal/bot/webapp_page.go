package bot

import (
	"net/http"
	"strings"
)

// webAppPage 返回 Mini App 单页(自包含,引 Telegram WebApp SDK)。
// __DEVPREVIEW__ 注入:仅 webapp_dev_preview=true 时允许从 ?initData= 读取(本地预览),生产为 false。
func (s *Service) webAppPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	flag := "false"
	if s.cfg.WebAppDevPreview {
		flag = "true"
	}
	_, _ = w.Write([]byte(strings.ReplaceAll(webAppHTML, "__DEVPREVIEW__", flag)))
}

// 主题取自 ../miaomiaowux/miaomiaowux-frontend/src/styles/theme.css:
// 品牌陶土橙 #d97757(暗色 #f18c6e),暖白底 / 深色 #10131c。logo 引主控 /images/。
const webAppHTML = `<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover">
<title>妙妙屋X · 我的面板</title>
<script src="https://telegram.org/js/telegram-web-app.js"></script>
<style>
:root{
 --brand:#d97757;--brand-soft:#fbe7da;--ok:#22c55e;--warn:#ef4444;--unknown:#a1a1aa;--radius:.625rem;
 --bg:#faf8f5;--text:#271610;--card:#ffffff;--muted:#816055;--border:rgba(137,110,96,.22);
}
html.dark{
 --brand:#f18c6e;--brand-soft:rgba(241,140,110,.16);
 --bg:#10131c;--text:#f9f4f1;--card:#1a1f2b;--muted:#cfb8af;--border:rgba(255,255,255,.08);
}
*{box-sizing:border-box}html,body{margin:0}
body{font-family:'Inter',-apple-system,system-ui,"PingFang SC","Microsoft YaHei",sans-serif;background:var(--bg);color:var(--text);padding-bottom:66px;}
header{position:sticky;top:0;z-index:10;background:var(--bg);border-bottom:1px solid var(--border);padding:10px 16px;display:flex;align-items:center;gap:10px;}
header .logo{height:26px;width:26px;display:block;flex:0 0 auto}
header .sub{font-size:11px;color:var(--muted);margin-left:auto}
.brand{font-size:18px;font-weight:700;display:inline-flex;align-items:baseline;line-height:1}
/* 特效 X(移植自 miaomiaowu-docs/src/components/animated-x.tsx) */
.ax{position:relative;display:inline-block;font-weight:800;margin-left:1px}
.ax-t{position:relative;z-index:10;background:linear-gradient(135deg,#f97316 0%,#ef4444 15%,#f59e0b 30%,#f97316 45%,#ec4899 60%,#f59e0b 75%,#f97316 90%,#ef4444 100%);background-size:300% 300%;-webkit-background-clip:text;background-clip:text;color:transparent;animation:axshift 3s ease-in-out infinite,axshimmer 2s ease-in-out infinite}
.ax-g{position:absolute;inset:0;z-index:0;opacity:.6;background:linear-gradient(135deg,#f97316,#ef4444,#f59e0b,#f97316);background-size:300% 300%;-webkit-background-clip:text;background-clip:text;color:transparent;animation:axshift 3s ease-in-out infinite,axglow 2s ease-in-out infinite}
.ax-p{position:absolute;inset:0;z-index:20;pointer-events:none}
.ax-p::before,.ax-p::after{content:'✦';position:absolute;font-size:.3em;animation:axspark 2.5s ease-in-out infinite}
.ax-p::before{top:-.2em;right:-.12em;color:#f59e0b;animation-delay:0s}
.ax-p::after{bottom:-.05em;left:-.12em;color:#f97316;animation-delay:1.2s}
@keyframes axshift{0%,100%{background-position:0% 50%}50%{background-position:100% 50%}}
@keyframes axshimmer{0%,100%{filter:brightness(1)}50%{filter:brightness(1.3)}}
@keyframes axglow{0%,100%{opacity:.3;filter:blur(8px)}50%{opacity:.7;filter:blur(12px)}}
@keyframes axspark{0%,100%{opacity:0;transform:scale(.5) rotate(0)}20%{opacity:1;transform:scale(1) rotate(90deg)}40%{opacity:0;transform:scale(.5) rotate(180deg)}}
main{padding:12px 14px}
.card{background:var(--card);border:1px solid var(--border);border-radius:var(--radius);padding:12px 14px;margin-bottom:10px;}
.title{font-size:12px;color:var(--muted);margin:0 0 7px;letter-spacing:.3px;}
.row{display:flex;justify-content:space-between;align-items:center;margin:4px 0;font-size:15px;}
.big{font-size:22px;font-weight:700;line-height:1.1}
.bar{height:8px;border-radius:6px;background:var(--brand-soft);overflow:hidden;margin:8px 0 0;}
.bar>i{display:block;height:100%;background:var(--brand);}
.muted{color:var(--muted);font-size:13px}
.sub{display:flex;justify-content:space-between;align-items:center;gap:8px;margin:10px 0}
.sub .u{font-size:12px;color:var(--muted);word-break:break-all;flex:1}
.btn{background:var(--brand);color:#fff;border:0;border-radius:8px;padding:7px 13px;font-size:13px;white-space:nowrap}
.inp{width:100%;margin:6px 0;padding:11px 12px;font-size:15px;border:1px solid var(--border);border-radius:8px;background:var(--bg);color:var(--text);outline:none}
.inp:focus{border-color:var(--brand)}
.seg{display:flex;gap:6px;margin:6px 0}
.seg button{flex:1;padding:8px;border:1px solid var(--border);border-radius:8px;background:var(--bg);color:var(--text);font-size:13px}
.seg button.active{background:var(--brand);color:#fff;border-color:var(--brand)}
.badge{font-size:11px;background:var(--brand);color:#fff;border-radius:6px;padding:1px 6px;margin-left:6px}
.warn{color:var(--warn)}
.dot{display:inline-block;width:9px;height:9px;border-radius:50%;margin-right:7px;vertical-align:middle}
.st{display:flex;align-items:center;justify-content:space-between;padding:9px 0;border-bottom:1px solid var(--border)}
.st:last-child{border-bottom:0}
.qm{display:inline-flex;width:9px;height:9px;margin-right:7px;align-items:center;justify-content:center;font-size:12px;font-weight:700;color:var(--unknown)}
.stlabel{font-size:12px}
#notice{text-align:center;color:var(--muted);padding:70px 24px;line-height:1.8}
.hide{display:none}
nav{position:fixed;left:0;right:0;bottom:0;display:flex;background:var(--card);border-top:1px solid var(--border);padding-bottom:env(safe-area-inset-bottom)}
nav button{flex:1;background:none;border:0;padding:8px 0 10px;color:var(--muted);display:flex;flex-direction:column;align-items:center;gap:3px;font-size:11px}
nav button.active{color:var(--brand)}
nav svg{width:22px;height:22px}
</style>
</head>
<body>
<header>
  <svg class="logo" viewBox="0 0 24 24" fill="none" stroke="var(--brand)" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-label="妙妙屋X"><path d="M15 6v12a3 3 0 1 0 3-3H6a3 3 0 1 0 3 3V6a3 3 0 1 0-3 3h12a3 3 0 1 0-3-3"></path></svg>
  <span class="brand">妙妙屋<span class="ax"><span class="ax-t">X</span><span class="ax-g" aria-hidden="true">X</span><span class="ax-p" aria-hidden="true"></span></span></span>
  <span class="sub">我的面板</span>
</header>
<div id="notice" class="hide">请在 Telegram 里通过机器人菜单按钮打开本页面。</div>
<main id="app" class="hide">
  <div id="view-home"></div>
  <div id="view-traffic" class="hide"></div>
  <div id="view-status" class="hide"></div>
  <div id="view-invites" class="hide"></div>
</main>
<nav id="nav" class="hide">
  <button data-v="home" class="active" onclick="__tab('home')">
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 10l9-7 9 7v9a2 2 0 0 1-2 2h-4v-6H9v6H5a2 2 0 0 1-2-2z"/></svg>
    <span>主页</span>
  </button>
  <button data-v="traffic" onclick="__tab('traffic')">
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19V5M4 19h16M8 19v-6M12 19V9M16 19v-9M20 19V7"/></svg>
    <span>流量</span>
  </button>
  <button data-v="status" onclick="__tab('status')">
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 12h4l3 8 4-16 3 8h4"/></svg>
    <span>状态</span>
  </button>
  <button id="nav-invites" data-v="invites" class="hide" onclick="__tab('invites')">
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16v16H4z"/><path d="M4 9h16M9 4v5"/></svg>
    <span>兑换码</span>
  </button>
</nav>
<script>
var tg = window.Telegram && window.Telegram.WebApp;
function hb(n){n=Number(n)||0;if(n<1024)return n+" B";var u=["KB","MB","GB","TB"],i=-1;do{n/=1024;i++}while(n>=1024&&i<3);return n.toFixed(2)+" "+u[i];}
function esc(s){return String(s==null?"":s).replace(/[&<>]/g,function(c){return{"&":"&amp;","<":"&lt;",">":"&gt;"}[c];});}
function setScheme(){var dark=tg?(tg.colorScheme==="dark"):window.matchMedia&&window.matchMedia("(prefers-color-scheme:dark)").matches;
 document.documentElement.classList.toggle("dark",!!dark);}
function copy(u){if(tg&&tg.HapticFeedback)tg.HapticFeedback.impactOccurred("light");
 navigator.clipboard&&navigator.clipboard.writeText(u);
 if(tg&&tg.showPopup)tg.showPopup({message:"订阅链接已复制"});else alert("已复制");}
window.__copy=copy;
function __tab(v){if(tg&&tg.HapticFeedback)tg.HapticFeedback.selectionChanged();
 ["home","traffic","status","invites"].forEach(function(x){document.getElementById("view-"+x).classList.toggle("hide",x!==v);});
 document.querySelectorAll("#nav button").forEach(function(b){b.classList.toggle("active",b.getAttribute("data-v")===v);});
 if(v==="invites"&&!window.__invLoaded){loadInvites();}}
window.__tab=__tab;

function chart(hist){
 hist=(hist||[]).filter(function(d){return d&&d.date;});
 if(!hist.length)return '<div class="card"><div class="title">每日流量</div><div class="muted">暂无数据</div></div>';
 var max=0;hist.forEach(function(d){if(d.used_gb>max)max=d.used_gb;});if(max<=0)max=1;
 var bw=Math.max(3,Math.floor((300-hist.length*2)/hist.length)),H=80,W=hist.length*(bw+2);
 var bars="";hist.forEach(function(d,i){var hh=Math.max(1,d.used_gb/max*H);bars+='<rect x="'+(i*(bw+2))+'" y="'+(H-hh)+'" width="'+bw+'" height="'+hh+'" rx="1" fill="var(--brand)"></rect>';});
 var total=hist.reduce(function(a,d){return a+(d.used_gb||0);},0);
 var h='<div class="card"><div class="title">每日流量(本周期)</div>';
 h+='<svg viewBox="0 0 '+W+' '+H+'" preserveAspectRatio="none" style="width:100%;height:80px">'+bars+'</svg>';
 h+='<div class="row"><span class="muted">'+esc(hist[0].date.slice(5))+' ~ '+esc(hist[hist.length-1].date.slice(5))+'</span><span class="muted">峰值 '+max.toFixed(2)+' GB · 累计 '+total.toFixed(2)+' GB</span></div></div>';
 return h;
}
function renderRegister(){
 var h='<div class="card"><div class="title">注册并绑定账号</div>';
 h+='<div class="muted" style="margin-bottom:10px">输入兑换码、设置用户名和密码,注册成功后自动绑定当前 Telegram。</div>';
 h+='<input id="r-code" class="inp" placeholder="兑换码" autocapitalize="characters">';
 h+='<input id="r-user" class="inp" placeholder="用户名(3-20 位,字母数字 _ -)">';
 h+='<input id="r-pwd" class="inp" type="password" placeholder="密码(6-64 位)">';
 h+='<div id="r-msg" class="warn" style="font-size:13px;min-height:18px;margin:4px 0"></div>';
 h+='<button class="btn" style="width:100%;padding:11px" onclick="__register(this)">注册并绑定</button>';
 h+='</div>';
 return h;
}
function renderRedeem(){
 var h='<div class="card"><div class="title">兑换码续期</div>';
 h+='<div style="display:flex;gap:8px"><input id="rd-code" class="inp" style="margin:0;flex:1" placeholder="输入兑换码" autocapitalize="characters">';
 h+='<button class="btn" onclick="__redeem(this)">兑换</button></div>';
 h+='<div id="rd-msg" style="font-size:13px;min-height:18px;margin:6px 0 0"></div>';
 h+='</div>';
 return h;
}
function __redeem(btn){
 var code=(document.getElementById("rd-code").value||"").trim();
 var msg=document.getElementById("rd-msg");msg.className="";msg.style.color="var(--muted)";
 if(!code){msg.style.color="var(--warn)";msg.textContent="请输入兑换码";return;}
 btn.disabled=true;btn.textContent="兑换中...";
 fetch("/api/tg-webapp/redeem",{method:"POST",headers:{"Content-Type":"application/json","X-Telegram-Init-Data":window.__init},body:JSON.stringify({code:code})})
  .then(function(r){return r.json().then(function(j){return{ok:r.ok,j:j};});})
  .then(function(res){btn.disabled=false;btn.textContent="兑换";
   if(!res.ok){msg.style.color="var(--warn)";msg.textContent=res.j.error||"兑换失败";return;}
   if(tg&&tg.HapticFeedback)tg.HapticFeedback.notificationOccurred("success");
   msg.style.color="var(--ok)";msg.textContent="续期成功,到期 "+(res.j.end_date||"");
   load();})
  .catch(function(){btn.disabled=false;btn.textContent="兑换";msg.style.color="var(--warn)";msg.textContent="网络错误";});
}
window.__redeem=__redeem;
function __register(btn){
 var code=(document.getElementById("r-code").value||"").trim();
 var user=(document.getElementById("r-user").value||"").trim();
 var pwd=(document.getElementById("r-pwd").value||"");
 var msg=document.getElementById("r-msg");
 if(!code||!user||!pwd){msg.textContent="请填写邀请码、用户名和密码";return;}
 if(!/^[a-zA-Z0-9_-]{3,20}$/.test(user)){msg.textContent="用户名格式不对(3-20 位 字母数字 _ -)";return;}
 if(pwd.length<6){msg.textContent="密码至少 6 位";return;}
 msg.textContent="";btn.disabled=true;btn.textContent="提交中...";
 fetch("/api/tg-webapp/register",{method:"POST",headers:{"Content-Type":"application/json","X-Telegram-Init-Data":window.__init},
  body:JSON.stringify({code:code,username:user,password:pwd})})
  .then(function(r){return r.json().then(function(j){return{ok:r.ok,j:j};});})
  .then(function(res){
   if(!res.ok){msg.textContent=res.j.error||"注册失败";btn.disabled=false;btn.textContent="注册并绑定";return;}
   if(tg&&tg.HapticFeedback)tg.HapticFeedback.notificationOccurred("success");
   load();
  })
  .catch(function(){msg.textContent="网络错误,请重试";btn.disabled=false;btn.textContent="注册并绑定";});
}
window.__register=__register;
function renderHome(d){
 if(d.bound===false)return renderRegister();
 var h="",a=d.account||{};
 h+='<div class="card"><div class="row" style="margin:0"><span style="font-size:17px;font-weight:600">'+esc(a.username)+'</span><span class="muted">'+(a.role==="admin"?"管理员":"用户")+(a.is_active?"":" · 已停用")+'</span></div>';
 if(a.email)h+='<div class="muted" style="margin-top:3px">'+esc(a.email)+'</div>';
 h+='</div>';
 var t=d.traffic;
 if(t){
  var used=Number(t.cycle_used)||0,limit=(Number(t.limit_gb)||0)*1073741824,pct=limit>0?Math.min(100,used/limit*100):0;
  var dl=t.days_left!=null?Number(t.days_left):null;
  h+='<div class="card"><div class="title">流量 · '+esc(t.package_name||"未绑定套餐")+'</div>';
  h+='<div class="row" style="margin:0"><span class="big">'+hb(used)+'</span><span class="muted">'+(limit>0?"/ "+(t.limit_gb)+" GB":"")+'</span></div>';
  if(limit>0){
   h+='<div class="bar"><i style="width:'+pct.toFixed(1)+'%"></i></div>';
   h+='<div class="row" style="margin:6px 0 0"><span class="muted">已用 '+pct.toFixed(1)+'%</span>'+(dl!=null?'<span class="'+(dl<=7?"warn":"muted")+'">剩 '+dl+' 天</span>':'<span></span>')+'</div>';
  }
  h+='<div class="row" style="margin:4px 0 0"><span class="muted">累计</span><span class="muted">↑'+hb(t.total_up)+' ↓'+hb(t.total_down)+'</span></div>';
  h+='</div>';
 }
 var subs=d.subscriptions||[];
 h+='<div class="card"><div class="title">订阅链接</div>';
 if(!subs.length)h+='<div class="muted">暂无订阅,请联系管理员。</div>';
 subs.forEach(function(sf){
  h+='<div class="sub"><div class="u"><div>'+esc(sf.name)+(sf["default"]?'<span class="badge">默认</span>':'')+'</div><div>'+esc(sf.url)+'</div></div>';
  h+='<button class="btn" onclick="__copy(\''+esc(sf.url)+'\')">复制</button></div>';
 });
 h+='</div>';
 h+=chart(d.history);
 h+=renderRedeem();
 return h;
}
function renderTraffic(d){
 var nodes=d.nodes||[];
 nodes=nodes.slice().sort(function(x,y){return (y.used||0)-(x.used||0);});
 var total=nodes.reduce(function(a,n){return a+(n.used||0);},0);
 var h='<div class="card"><div class="title">各节点已用流量 · 合计 '+hb(total)+'</div>';
 if(!nodes.length)h+='<div class="muted">本周期暂无各节点流量数据。</div>';
 var max=0;nodes.forEach(function(n){if((n.used||0)>max)max=n.used;});if(max<=0)max=1;
 nodes.forEach(function(n){var pct=Math.max(2,(n.used||0)/max*100);
  h+='<div style="margin:11px 0"><div class="row" style="margin:0 0 5px"><span>'+esc(n.name)+'</span><span class="muted">'+hb(n.used)+'</span></div>';
  h+='<div class="bar" style="margin:0"><i style="width:'+pct+'%"></i></div></div>';});
 h+='</div>';
 return h;
}
function renderStatus(d){
 var ns=d.node_status||[];
 var on=ns.filter(function(n){return n.status==="online";}).length;
 var h='<div class="card"><div class="title">节点状态 · 在线 '+on+'/'+ns.length+'</div>';
 if(!ns.length)h+='<div class="muted">套餐内暂无节点。</div>';
 ns.forEach(function(n){var ic,lb,col;
  if(n.status==="online"){ic='<span class="dot" style="background:var(--ok)"></span>';lb="在线";col="var(--ok)";}
  else if(n.status==="unknown"){ic='<span class="qm">?</span>';lb="外部";col="var(--unknown)";}
  else{ic='<span class="dot" style="background:var(--warn)"></span>';lb="离线";col="var(--warn)";}
  h+='<div class="st"><span>'+ic+esc(n.name)+' <span class="muted">'+esc(n.protocol||"")+'</span></span><span class="stlabel" style="color:'+col+'">'+lb+'</span></div>';});
 h+='</div>';
 return h;
}
function render(d){
 document.getElementById("app").classList.remove("hide");
 document.getElementById("view-home").innerHTML=renderHome(d);
 document.getElementById("view-traffic").innerHTML=renderTraffic(d);
 document.getElementById("view-status").innerHTML=renderStatus(d);
 if(d.bound!==false)document.getElementById("nav").classList.remove("hide");
 if(d.is_admin)document.getElementById("nav-invites").classList.remove("hide");
}

// ===== 管理员:邀请码 =====
function invStatus(ic){if(ic.revoked)return["✗ 已撤销","var(--warn)"];if(!ic.usable)return["○ 已用尽/过期","var(--unknown)"];return["✓ 可用","var(--ok)"];}
function renderInvites(){
 var inv=window.__inv||{},pkgs=inv.packages||[],list=inv.invites||[];
 var h='<div class="card"><div class="title">生成兑换码</div>';
 h+='<div class="muted" style="margin-bottom:4px">套餐</div>';
 h+='<select id="i-pkg" class="inp" style="margin-top:0">';
 h+='<option value="0">不绑套餐</option>';
 pkgs.forEach(function(p){h+='<option value="'+p.id+'">'+esc(p.name)+'('+(p.traffic_limit_gb)+'GB/'+(p.cycle_days)+'天)</option>';});
 h+='</select>';
 h+='<div class="muted" style="margin:10px 0 4px">有效期</div><div class="seg" id="i-dur">';
 [1,3,6,12].forEach(function(m,i){h+='<button class="'+(i==0?"active":"")+'" onclick="__durpick(this,'+m+')">'+(m==12?"1年":m+"月")+'</button>';});
 h+='</div>';
 h+='<div id="i-msg" class="warn" style="font-size:13px;min-height:18px;margin:4px 0"></div>';
 h+='<button class="btn" style="width:100%;padding:11px" onclick="__createInv(this)">生成兑换码</button>';
 h+='</div>';
 // list
 h+='<div class="card"><div class="title">兑换码('+list.length+')</div>';
 if(!list.length)h+='<div class="muted">暂无兑换码。</div>';
 list.forEach(function(ic){var st=invStatus(ic);
  h+='<div class="st"><div style="flex:1"><div style="font-family:monospace">'+esc(ic.code)+'</div>';
  h+='<div class="muted" style="font-size:12px">'+esc(st[0])+' · '+ic.used_count+'/'+ic.max_uses+'</div></div>';
  if(!ic.revoked)h+='<button class="btn" style="background:var(--warn)" onclick="__revokeInv(\''+esc(ic.code)+'\',this)">撤销</button>';
  h+='</div>';});
 h+='</div>';
 document.getElementById("view-invites").innerHTML=h;
 window.__dur=1;
}
function __durpick(btn,m){window.__dur=m;document.querySelectorAll("#i-dur button").forEach(function(b){b.classList.remove("active");});btn.classList.add("active");}
window.__durpick=__durpick;
function loadInvites(){
 document.getElementById("view-invites").innerHTML='<div class="card"><div class="muted">加载中...</div></div>';
 fetch("/api/tg-webapp/admin/invites",{headers:{"X-Telegram-Init-Data":window.__init}})
  .then(function(r){if(!r.ok)throw new Error(r.status);return r.json();})
  .then(function(j){window.__inv=j;window.__invLoaded=true;renderInvites();})
  .catch(function(){document.getElementById("view-invites").innerHTML='<div class="card">加载失败。</div>';});
}
function __createInv(btn){
 var msg=document.getElementById("i-msg");msg.textContent="";
 var pkg=parseInt(document.getElementById("i-pkg").value,10)||0;
 var body={package_id:pkg?pkg:null,duration_months:window.__dur||1};
 btn.disabled=true;btn.textContent="生成中...";
 fetch("/api/tg-webapp/admin/invite-create",{method:"POST",headers:{"Content-Type":"application/json","X-Telegram-Init-Data":window.__init},body:JSON.stringify(body)})
  .then(function(r){return r.json().then(function(j){return{ok:r.ok,j:j};});})
  .then(function(res){btn.disabled=false;btn.textContent="生成兑换码";
   if(!res.ok){msg.textContent=res.j.error||"生成失败";return;}
   if(tg&&tg.HapticFeedback)tg.HapticFeedback.notificationOccurred("success");
   if(tg&&tg.showPopup)tg.showPopup({title:"已生成",message:"兑换码:"+res.j.code});
   loadInvites();})
  .catch(function(){btn.disabled=false;btn.textContent="生成兑换码";msg.textContent="网络错误";});
}
window.__createInv=__createInv;
function __revokeInv(code,btn){
 btn.disabled=true;btn.textContent="...";
 fetch("/api/tg-webapp/admin/invite-revoke",{method:"POST",headers:{"Content-Type":"application/json","X-Telegram-Init-Data":window.__init},body:JSON.stringify({code:code})})
  .then(function(r){return r.json().then(function(j){return{ok:r.ok,j:j};});})
  .then(function(res){if(res.ok){if(tg&&tg.HapticFeedback)tg.HapticFeedback.impactOccurred("medium");loadInvites();}else{btn.disabled=false;btn.textContent="撤销";if(tg&&tg.showPopup)tg.showPopup({message:res.j.error||"撤销失败"});}})
  .catch(function(){btn.disabled=false;btn.textContent="撤销";});
}
window.__revokeInv=__revokeInv;
function load(){
 fetch("/api/tg-webapp/me",{headers:{"X-Telegram-Init-Data":window.__init}})
  .then(function(r){if(!r.ok)throw new Error(r.status);return r.json();})
  .then(render)
  .catch(function(){document.getElementById("app").classList.remove("hide");document.getElementById("view-home").innerHTML='<div class="card">加载失败,请稍后重试。</div>';});
}
(function(){
 setScheme();
 if(tg&&tg.onEvent)tg.onEvent("themeChanged",setScheme);
 var initData=(tg&&tg.initData)?tg.initData:(__DEVPREVIEW__?new URLSearchParams(location.search).get("initData"):"");
 if(!initData){document.getElementById("notice").classList.remove("hide");return;}
 window.__init=initData;
 if(tg){tg.ready();tg.expand();}
 load();
})();
</script>
</body>
</html>`
