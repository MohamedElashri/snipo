const {test,expect}=require('@playwright/test');
const fs=require('node:fs');
const vm=require('node:vm');

test('context menu reuses the content script before trying injection',async()=>{
  let clicked;
  let injected=0;
  const messages=[];
  const chrome={runtime:{onInstalled:{addListener(){}},onMessage:{addListener(){}}},
    contextMenus:{onClicked:{addListener(fn){clicked=fn}}},
    tabs:{sendMessage:async(id,message)=>{messages.push(message)}},
    scripting:{insertCSS:async()=>{injected++},executeScript:async()=>{injected++}}};
  const source=fs.readFileSync(require.resolve('../background.src.js'),'utf8').replace(/^import .*;\n/,'');
  vm.runInNewContext(source,{chrome,console,URL,SnipoClient:class{}});
  await clicked({menuItemId:'save-to-snipo',selectionText:'code'},{id:1,title:'Example'});
  expect(injected).toBe(0);
  expect(messages).toHaveLength(1);
  expect(messages[0].action).toBe('contextMenuSave');
});
