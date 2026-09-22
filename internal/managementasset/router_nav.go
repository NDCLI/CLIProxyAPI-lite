package managementasset

import "bytes"

// AddRouterNavigation presents the working management routes in a compact rail.
// Features without a gateway implementation are clearly marked unavailable.
func AddRouterNavigation(html []byte) ([]byte, bool) {
	if !bytes.Contains(html, []byte("</head>")) || bytes.Contains(html, []byte("cliproxy-router-nav")) {
		return html, false
	}
	const extension = `<style>
#cliproxy-router-nav{display:none;padding:8px;border-bottom:1px solid rgba(255,255,255,.1);font-family:inherit}
#cliproxy-router-nav .router-nav-item{display:flex;align-items:center;gap:12px;width:100%;min-height:34px;box-sizing:border-box;margin:2px 0;padding:7px 12px;border:0;border-radius:8px;background:transparent;color:#a7adb8;text-align:left;font-family:inherit;font-size:13px;font-weight:500;line-height:1.2;cursor:pointer}
#cliproxy-router-nav .router-nav-item:hover{background:#332b29;color:#f5f5f5}
#cliproxy-router-nav .router-nav-item.active{background:#3b2c29;color:#ff7658}
#cliproxy-router-nav .router-nav-item:disabled{opacity:.48;cursor:not-allowed}
#cliproxy-router-nav .router-nav-item:disabled:hover{background:transparent;color:#a7adb8}
#cliproxy-router-nav svg{flex:none;width:18px;height:18px;fill:none;stroke:currentColor;stroke-width:1.8;stroke-linecap:round;stroke-linejoin:round}
#cliproxy-router-nav .router-nav-item span{min-width:0}
#cliproxy-router-nav .router-nav-more{width:100%;margin-top:9px;padding:9px 12px;border:0;border-top:1px solid rgba(255,255,255,.1);background:transparent;color:#a7adb8;text-align:left;font:inherit;font-size:12px;cursor:pointer}
#cliproxy-router-nav .router-nav-more:hover{color:#fff}
#cliproxy-router-nav:not(.expanded) ~ *{display:none!important}
</style>
<script>
(()=>{
const icons={endpoint:'<path d="M4 7h16M4 12h16M4 17h16"/><circle cx="8" cy="7" r="1"/><circle cx="15" cy="12" r="1"/>',providers:'<rect x="4" y="4" width="16" height="7" rx="1"/><rect x="4" y="13" width="16" height="7" rx="1"/>',combo:'<path d="m12 3 9 6-9 6-9-6 9-6Zm-9 12 9 6 9-6"/>',usage:'<path d="M4 20V12m5 8V5m5 15v-9m5 9V8"/>',quota:'<path d="M12 3a9 9 0 1 0 9 9h-9V3Z"/><path d="M15 3a9 9 0 0 1 6 6h-6V3Z"/>',token:'<path d="M12 2 4 13h6l-1 9 11-13h-6l1-7Z"/>',cli:'<rect x="3" y="4" width="18" height="16" rx="2"/><path d="m7 9 3 3-3 3m5 0h5"/>'};
const items=[
{id:'endpoint',label:'Endpoint & Key',path:'/config'},
{id:'providers',label:'Providers',path:'/ai-providers'},
{id:'combo',label:'Combo & Vision Adapter'},
{id:'usage',label:'Usage',path:'/usage-history'},
{id:'quota',label:'Quota Tracker',path:'/quota'},
{id:'token',label:'Token Saver'},
{id:'cli',label:'CLI Tools'}];
function mount(){
  if(document.getElementById('cliproxy-router-nav'))return;
  const anchor=[...document.querySelectorAll('a')].find(a=>/ai-providers|quick-start|\/config/.test(a.getAttribute('href')||''));
  const nav=anchor?.closest('nav')||anchor?.parentElement?.parentElement;
  if(!nav)return;
  const rail=document.createElement('div');rail.id='cliproxy-router-nav';
  rail.innerHTML=items.map(i=>'<button type="button" class="router-nav-item" data-router-id="'+i.id+'"'+(i.path?'':' disabled title="Chưa có xử lý ở gateway"')+'><svg viewBox="0 0 24 24" aria-hidden="true">'+icons[i.id]+'</svg><span>'+i.label+'</span></button>').join('')+'<button type="button" class="router-nav-more">Các mục quản lý khác ▾</button>';
  nav.insertAdjacentElement('afterbegin',rail);
  let expanded=false;
  function showOriginal(){rail.classList.toggle('expanded',expanded);rail.querySelector('.router-nav-more').textContent=expanded?'Thu gọn các mục quản lý ▴':'Các mục quản lý khác ▾'}
  rail.querySelector('.router-nav-more').onclick=()=>{expanded=!expanded;showOriginal()};
  rail.querySelectorAll('[data-router-id]').forEach(button=>button.onclick=()=>{const item=items.find(i=>i.id===button.dataset.routerId);if(item?.path)location.hash='#'+item.path});
  rail.style.display='block';showOriginal();sync();
}
function sync(){const path=location.hash.replace(/^#/, '').split('?')[0];document.querySelectorAll('#cliproxy-router-nav [data-router-id]').forEach(button=>{const item=items.find(i=>i.id===button.dataset.routerId);button.classList.toggle('active',!!item?.path&&path===item.path)})}
window.addEventListener('hashchange',sync);
const timer=setInterval(()=>{mount();if(document.getElementById('cliproxy-router-nav'))clearInterval(timer)},500);
if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',mount);else mount();
})()
</script>`
	return bytes.Replace(html, []byte("</head>"), append([]byte(extension), []byte("</head>")...), 1), true
}
