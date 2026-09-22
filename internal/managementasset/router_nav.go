package managementasset

import "bytes"

// AddRouterNavigation adds an experimental 9router-style navigation rail.
// Existing management routes remain the source of truth for supported pages.
func AddRouterNavigation(html []byte) ([]byte, bool) {
	if !bytes.Contains(html, []byte("</head>")) || bytes.Contains(html, []byte("cliproxy-router-nav")) {
		return html, false
	}
	const extension = `<style>
#cliproxy-router-nav{display:none;position:relative;z-index:5;padding:10px 8px 12px;margin:0 8px;border-bottom:1px solid rgba(255,255,255,.08);font-family:inherit}#cliproxy-router-nav .router-nav-title{display:none}#cliproxy-router-nav .router-nav-item{display:flex;align-items:center;gap:12px;width:100%;box-sizing:border-box;border:0;border-radius:8px;padding:8px 12px;margin:2px 0;background:transparent;color:#a7adb8;text-align:left;font:500 13px/1.2 inherit;cursor:pointer;transition:background-color .15s,color .15s}#cliproxy-router-nav .router-nav-item:hover{background:#332b29;color:#f5f5f5}#cliproxy-router-nav .router-nav-item.active{background:#3b2c29;color:#ff7658}#cliproxy-router-nav .router-nav-item .material-symbols-outlined{font-size:19px;line-height:1;color:currentColor}#cliproxy-router-nav .router-nav-item small{display:block;margin-left:auto;color:#777;font-size:10px;font-weight:400}.cliproxy-router-panel{display:none;box-sizing:border-box;max-width:1120px;margin:0 auto;padding:28px;color:inherit;font:inherit}.cliproxy-router-panel h1{margin:0 0 8px;font-size:26px}.cliproxy-router-panel p{color:#9ca3af}.cliproxy-router-panel .router-card-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:14px;margin-top:24px}.cliproxy-router-panel .router-card{padding:18px;border:1px solid rgba(255,255,255,.1);border-radius:12px;background:rgba(31,31,31,.72)}.cliproxy-router-panel .router-card h2{margin:0 0 8px;font-size:15px}.cliproxy-router-panel .router-card button{margin-top:12px;border:1px solid #65534b;border-radius:7px;padding:7px 10px;background:#3b2c29;color:#ff9278;cursor:pointer;font:inherit}.cliproxy-router-panel .router-toggle{display:flex;align-items:center;gap:10px;margin-top:14px;color:#d1d5db}.cliproxy-router-panel input[type=checkbox]{accent-color:#f27657}@media(max-width:900px){#cliproxy-router-nav{margin:0}.cliproxy-router-panel .router-card-grid{grid-template-columns:1fr}}
</style>
<script>
(()=>{const items=[
{id:'endpoint',label:'Endpoint & Key',icon:'api',find:/Quick Start|Config Panel|Endpoint/i},
{id:'providers',label:'Providers',icon:'dns',find:/AI Providers|Providers/i},
{id:'combo',label:'Combo & Vision Adapter',icon:'layers',custom:'combo'},
{id:'usage',label:'Usage',icon:'bar_chart',custom:'usage'},
{id:'quota',label:'Quota Tracker',icon:'data_usage',find:/Quota Management/i},
{id:'token',label:'Token Saver',icon:'savings',custom:'token'},
{id:'cli',label:'CLI Tools',icon:'terminal',custom:'cli'}];
let active='',original=[],panel=null;
const textOf=e=>(e?.textContent||'').replace(/\s+/g,' ').trim();
function targetFor(item){return [...document.querySelectorAll('a,button')].find(e=>e!==document.getElementById('cliproxy-usage-history-link')&&item.find?.test(textOf(e)))}
function main(){return document.querySelector('main')||document.body}
function hide(){const root=main();original=[...root.children].filter(e=>e.id!=='cliproxy-router-panel'&&e.id!=='cliproxy-usage-history');original.forEach(e=>{e.dataset.routerDisplay=e.style.display;e.style.display='none'})}
function restore(){original.forEach(e=>{e.style.display=e.dataset.routerDisplay||'';delete e.dataset.routerDisplay});original=[];panel?.remove();panel=null}
function panelFor(id){const content={combo:['Combo & Vision Adapter','Tạo nhóm model theo thứ tự ưu tiên và chuyển dự phòng khi model chính lỗi.','<div class="router-card-grid"><div class="router-card"><h2>Model combo</h2><p>Thử nghiệm: cấu hình model chính và model dự phòng.</p><label class="router-toggle"><input type="checkbox" checked> Bật chuyển dự phòng</label></div><div class="router-card"><h2>Vision adapter</h2><p>Định tuyến request hình ảnh đến model có khả năng vision.</p><label class="router-toggle"><input type="checkbox"> Tự động chọn vision</label></div></div>'],token:['Token Saver','Tối ưu prompt trước khi chuyển đến provider. Tính năng này đang ở chế độ thử nghiệm.','<div class="router-card-grid"><div class="router-card"><h2>Token Saver</h2><p>Chưa bật xử lý tự động trên backend.</p><label class="router-toggle"><input type="checkbox"> Bật thử nghiệm</label></div></div>'],cli:['CLI Tools','Cấu hình nhanh các công cụ CLI dùng chung endpoint và API key.','<div class="router-card-grid"><div class="router-card"><h2>Claude / Codex / OpenCode</h2><p>Dùng endpoint hiện tại để kết nối công cụ CLI tương thích.</p><button type="button">Mở hướng dẫn cấu hình</button></div></div>']}[id];if(!content)return;panel=document.createElement('section');panel.id='cliproxy-router-panel';panel.className='cliproxy-router-panel';panel.innerHTML='<h1>'+content[0]+'</h1><p>'+content[1]+'</p>'+content[2];main().appendChild(panel)}
function go(item){active=item.id;document.querySelectorAll('#cliproxy-router-nav .router-nav-item').forEach(e=>e.classList.toggle('active',e.dataset.id===active));if(item.custom==='usage'){restore();location.hash='#/usage-history';window.dispatchEvent(new HashChangeEvent('hashchange'));return}if(item.custom){hide();panelFor(item.custom);return}restore();const target=targetFor(item);if(target)target.click()}
function nav(){if(document.getElementById('cliproxy-router-nav'))return;const anchors=[...document.querySelectorAll('a,button')];const anchor=anchors.find(e=>/AI Providers|Providers|Quick Start|Config Panel/i.test(textOf(e)));if(!anchor)return;const host=anchor.closest('nav')||anchor.parentElement?.parentElement||anchor.parentElement;if(!host)return;const bar=document.createElement('div');bar.id='cliproxy-router-nav';bar.innerHTML=items.map(i=>'<button type="button" class="router-nav-item" data-id="'+i.id+'"><span class="material-symbols-outlined">'+i.icon+'</span><span>'+i.label+'</span></button>').join('');host.insertAdjacentElement('afterbegin',bar);bar.querySelectorAll('button').forEach(b=>b.onclick=()=>go(items.find(i=>i.id===b.dataset.id)));bar.style.display='block'}
function sync(){nav();const activeHash=location.hash.includes('usage-history');if(activeHash){document.querySelector('#cliproxy-router-nav [data-id="usage"]')?.classList.add('active')}}
window.addEventListener('hashchange',sync);setInterval(sync,700);if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',sync);else sync()})()
</script>`
	return bytes.Replace(html, []byte("</head>"), append([]byte(extension), []byte("</head>")...), 1), true
}
