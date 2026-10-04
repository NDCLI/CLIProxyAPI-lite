// Run with Node.js and an existing Playwright installation (NODE_PATH if needed).
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const {chromium} = require('playwright');
(async () => {
 const browser = await chromium.launch({channel:'msedge',headless:true});
 try {
 const page = await browser.newPage();
 await page.setContent('<div id="page"></div>');
 const script = fs.readFileSync(path.join(__dirname, '../internal/managementasset/web/app.js'),'utf8');
 const dirty = script.slice(script.indexOf('function bindDirtyAction('),script.indexOf('function endpointKeyTable('));
 const controls = script.slice(script.indexOf('function compatibleProviderCard('),script.indexOf('function providerAccountRow('));
 await page.addScriptTag({content: `const state={locale:'en'}; const t=k=>k; const icon=()=>''; const escapeHTML=v=>String(v); const matchMedia=()=>({matches:true}); ${dirty}\n${controls}`});
 await page.evaluate(() => {
  window.items=[{id:'test',name:'Test',base_url:'https://example.test/v1',enabled:true,api_key_masks:['first-mask','second-mask'],api_key_enabled:[true,true],models:[]}];
  window.calls=[];window.fail=false;
  window.api=async (url,options)=> {const body=JSON.parse(options.body);calls.push(body); if(fail) throw new Error('failed');items[0].api_key_enabled[body.api_key_index]=body.api_key_action==='enable';return {item:structuredClone(items[0])};};
  const container=document.getElementById('page'); container.innerHTML=compatibleProviderSection(items); bindCompatibleProviderControls(container,items,()=>{});
 });
 await page.locator('[data-custom-edit]').click();
 await page.locator('[name="prefix"]').fill('unsaved-prefix');
 await page.locator('[data-provider-api-key-toggle="0"]').click();
 await page.waitForFunction(()=>document.querySelector('[data-provider-api-key-toggle="0"]').dataset.enabled==='false');
 assert.deepEqual(await page.evaluate(()=>calls[0]),{api_key_action:'disable',api_key_index:0});
 assert.equal(await page.locator('[data-provider-api-key-toggle="1"]').getAttribute('data-enabled'),'true');
 assert.equal(await page.locator('[name="prefix"]').inputValue(),'unsaved-prefix');
 await page.locator('[data-provider-api-key-toggle="0"]').click();
 await page.waitForFunction(()=>document.querySelector('[data-provider-api-key-toggle="0"]').dataset.enabled==='true');
 await page.evaluate(()=>{window.fail=true;});
 await page.locator('[data-provider-api-key-toggle="1"]').click();
 await page.locator('[data-provider-api-key-result]').nth(1).waitFor({state:'visible'});
 assert.equal(await page.locator('[data-provider-api-key-toggle="1"]').getAttribute('data-enabled'),'true');
 assert.equal(await page.locator('[data-provider-api-key-toggle="1"]').isEnabled(),true);
 console.log('PASS: per-key toggle, sibling state, draft preservation, reenable and failed save');
 } finally {await browser.close();}
})().catch(err=>{console.error(err);process.exitCode=1;});
