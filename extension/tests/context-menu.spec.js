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

test('a browser with existing scripting access can initialize a missing content script',async()=>{
  let clicked;
  let ready=false;
  let injected=0;
  const messages=[];
  const chrome={runtime:{onInstalled:{addListener(){}},onMessage:{addListener(){}}},
    contextMenus:{onClicked:{addListener(fn){clicked=fn}}},
    tabs:{sendMessage:async(id,message)=>{if(!ready)throw new Error('No receiver');messages.push(message)}},
    scripting:{insertCSS:async()=>{injected++},executeScript:async()=>{injected++;ready=true}}};
  const source=fs.readFileSync(require.resolve('../background.src.js'),'utf8').replace(/^import .*;\n/,'');
  vm.runInNewContext(source,{chrome,console,URL,SnipoClient:class{}});
  await clicked({menuItemId:'save-to-snipo',selectionText:'code'},{id:1,title:'Example'});
  expect(injected).toBe(2);
  expect(messages).toHaveLength(1);
  expect(messages[0].action).toBe('contextMenuSave');
});

test('a browser without scripting access safely quick-saves when no page receiver exists',async()=>{
  let clicked;
  const saved=[];
  const chrome={runtime:{onInstalled:{addListener(){}},onMessage:{addListener(){}}},
    contextMenus:{onClicked:{addListener(fn){clicked=fn}}},
    storage:{sync:{get:async()=>({instanceUrl:'https://snipo.test',apiKey:'test'})}},
    tabs:{sendMessage:async(id,message)=>{if(message.action==='contextMenuSave')throw new Error('No receiver')}}};
  class Client {async createSnippet(payload){saved.push(payload);return {id:'saved'}}}
  const source=fs.readFileSync(require.resolve('../background.src.js'),'utf8').replace(/^import .*;\n/,'');
  vm.runInNewContext(source,{chrome,console,URL,SnipoClient:Client});
  await clicked({menuItemId:'save-to-snipo',selectionText:'    code\n'},{id:1,title:'Example'});
  expect(saved).toHaveLength(1);
  expect(saved[0].files[0].content).toBe('    code\n');
});
