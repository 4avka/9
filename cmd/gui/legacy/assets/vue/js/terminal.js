!function(t,n){"object"==typeof exports&&"object"==typeof module?module.exports=n():"function"==typeof define&&define.amd?define("VueTerminal",[],n):"object"==typeof exports?exports.VueTerminal=n():t.VueTerminal=n()}("undefined"!=typeof self?self:this,function(){return function(t){function n(i){if(e[i])return e[i].exports;var r=e[i]={i:i,l:!1,exports:{}};return t[i].call(r.exports,r,r.exports,n),r.l=!0,r.exports}var e={};return n.m=t,n.c=e,n.d=function(t,e,i){n.o(t,e)||Object.defineProperty(t,e,{configurable:!1,enumerable:!0,get:i})},n.n=function(t){var e=t&&t.__esModule?function(){return t.default}:function(){return t};return n.d(e,"a",e),e},n.o=function(t,n){return Object.prototype.hasOwnProperty.call(t,n)},n.p="/dist/",n(n.s=10)}([function(t,n){var e=t.exports={version:"2.6.2"};"number"==typeof __e&&(__e=e)},function(t,n){var e=t.exports="undefined"!=typeof window&&window.Math==Math?window:"undefined"!=typeof self&&self.Math==Math?self:Function("return this")();"number"==typeof __g&&(__g=e)},function(t,n){t.exports=function(t){return"object"==typeof t?null!==t:"function"==typeof t}},function(t,n,e){t.exports=!e(4)(function(){return 7!=Object.defineProperty({},"a",{get:function(){return 7}}).a})},function(t,n){t.exports=function(t){try{return!!t()}catch(t){return!0}}},function(t,n,e){"use strict";var i=e(18),r=e.n(i);n.a={name:"VueTerminal",data:function(){return{title:"vTerminal",messageList:[],actionResult:"",lastLineContent:"...",inputCommand:"",supportingCommandList:"",historyIndex:0,commandHistory:[]}},props:{defaultTask:{required:!1,default:"defaultTask"},commandList:{required:!1,default:function(){return{}}},taskList:{required:!1,default:function(){return{}}}},computed:{lastLineClass:function(){return"&nbsp"===this.lastLineContent?"cursor":"..."===this.lastLineContent?"loading":void 0}},created:function(){var t=this;this.supportingCommandList=r()(this.commandList).concat(r()(this.taskList)),this.handleRun(this.defaultTask).then(function(){t.pushToList({level:"System",message:'Type "help" to get a supporting command list.'}),t.handleFocus()})},methods:{handleFocus:function(){this.$refs.inputBox.focus()},handleCommand:function(t){var n=this;if(13!==t.keyCode)return void this.handlekeyEvent(t);if(this.commandHistory.push(this.inputCommand),this.historyIndex=this.commandHistory.length,this.pushToList({message:"$ \\ "+this.title+" "+this.inputCommand+" "}),this.inputCommand){var e=this.inputCommand.split(" ");"help"===e[0]?this.printHelp(e[1]):this.commandList[this.inputCommand]?this.commandList[this.inputCommand].messages.map(function(t){return n.pushToList(t)}):this.taskList[this.inputCommand.split(" ")[0]]?this.handleRun(this.inputCommand.split(" ")[0],this.inputCommand):(this.pushToList({level:"System",message:"Unknown Command."}),this.pushToList({level:"System",message:'type "help" to get a supporting command list.'})),this.inputCommand="",this.autoScroll()}},handlekeyEvent:function(t){switch(t.keyCode){case 38:this.historyIndex=0===this.historyIndex?0:this.historyIndex-1,this.inputCommand=this.commandHistory[this.historyIndex];break;case 40:this.historyIndex=this.historyIndex===this.commandHistory.length?this.commandHistory.length:this.historyIndex+1,this.inputCommand=this.commandHistory[this.historyIndex]}},handleRun:function(t,n){var e=this;return this.lastLineContent="...",this.taskList[t][t](this.pushToList,n).then(function(t){e.pushToList(t),e.lastLineContent="&nbsp"}).catch(function(t){e.pushToList(t||{type:"error",label:"Error",message:"Something went wrong!"}),e.lastLineContent="&nbsp"})},pushToList:function(t){this.messageList.push(t),this.autoScroll()},printHelp:function(t){var n=this;if(t){var e=this.commandList[t]||this.taskList[t];this.pushToList({message:e.description})}else this.pushToList({message:"Here is a list of supporting command."}),this.supportingCommandList.map(function(t){n.commandList[t]?n.pushToList({type:"success",label:t,message:"---\x3e "+n.commandList[t].description}):n.pushToList({type:"success",label:t,message:"---\x3e "+n.taskList[t].description})}),this.pushToList({message:"Enter help <command> to get help for a particular command."});this.autoScroll()},time:function(){return(new Date).toLocaleTimeString().split("").splice(2).join("")},autoScroll:function(){var t=this;this.$nextTick(function(){t.$refs.terminalWindow.scrollTop=t.$refs.terminalLastLine.offsetTop})}}}},function(t,n){t.exports=function(t){if(void 0==t)throw TypeError("Can't call method on  "+t);return t}},function(t,n){var e={}.hasOwnProperty;t.exports=function(t,n){return e.call(t,n)}},function(t,n,e){var i=e(24),r=e(6);t.exports=function(t){return i(r(t))}},function(t,n){var e=Math.ceil,i=Math.floor;t.exports=function(t){return isNaN(t=+t)?0:(t>0?i:e)(t)}},function(t,n,e){"use strict";Object.defineProperty(n,"__esModule",{value:!0});var i=e(11);n.default=i.a,"undefined"!=typeof window&&window.Vue&&window.Vue.component("vue-terminal",i.a)},function(t,n,e){"use strict";function i(t){e(12)}var r=e(5),o=e(45),a=e(17),s=i,u=a(r.a,o.a,!1,s,"data-v-58447793",null);n.a=u.exports},function(t,n,e){var i=e(13);"string"==typeof i&&(i=[[t.i,i,""]]),i.locals&&(t.exports=i.locals);e(15)("630c3f82",i,!0,{})},function(t,n,e){n=t.exports=e(14)(!1),n.push([t.i,'.terminal[data-v-58447793]{position:relative;width:100%;border-radius:4px;color:#fff;margin-bottom:10px;max-height:580px}.terminal .terminal-window[data-v-58447793]{padding-top:50px;background-color:#030924;min-height:140px;padding:20px;font-weight:400;font-family:Monaco,Menlo,Consolas,monospace;color:#fff}.terminal .terminal-window pre[data-v-58447793]{font-family:Monaco,Menlo,Consolas,monospace;white-space:pre-wrap}.terminal .terminal-window p[data-v-58447793]{overflow-wrap:break-word;word-break:break-all;font-size:13px}.terminal .terminal-window p .cmd[data-v-58447793]{line-height:24px}.terminal .terminal-window p .info[data-v-58447793]{padding:2px 3px;background:#2980b9}.terminal .terminal-window p .warning[data-v-58447793]{padding:2px 3px;background:#f39c12}.terminal .terminal-window p .success[data-v-58447793]{padding:2px 3px;background:#27ae60}.terminal .terminal-window p .error[data-v-58447793]{padding:2px 3px;background:#c0392b}.terminal .terminal-window p .system[data-v-58447793]{padding:2px 3px;background:#bdc3c7}.terminal .terminal-window pre[data-v-58447793]{display:inline}.terminal .header ul.shell-dots li[data-v-58447793]{display:inline-block;width:12px;height:12px;border-radius:6px;background-color:#030924;margin-left:6px}.terminal .header ul.shell-dots li.red[data-v-58447793]{background-color:#c83030}.terminal .header ul.shell-dots li.yellow[data-v-58447793]{background-color:#f7db60}.terminal .header ul.shell-dots li.green[data-v-58447793]{background-color:#2ec971}.terminal .header[data-v-58447793]{position:absolute;z-index:2;top:0;right:0;left:0;background-color:#959598;text-align:center;padding:2px;border-top-left-radius:4px;border-top-right-radius:4px}.terminal .header h4[data-v-58447793]{font-size:14px;margin:5px;letter-spacing:1px}.terminal .header ul.shell-dots[data-v-58447793]{position:absolute;top:5px;left:8px;padding-left:0;margin:0}.terminal .terminal-window .prompt[data-v-58447793]:before{content:"$";margin-right:10px}.terminal .terminal-window .cursor[data-v-58447793]{margin:0;background-color:#fff;animation:blink-data-v-58447793 1s step-end infinite;-webkit-animation:blink-data-v-58447793 1s step-end infinite;margin-left:-5px}@keyframes blink-data-v-58447793{50%{visibility:hidden}}@-webkit-keyframes blink-data-v-58447793{50%{visibility:hidden}}.terminal .terminal-window .loading[data-v-58447793]{display:inline-block;width:0;overflow:hidden;animation:load-data-v-58447793 1.2s step-end infinite;-webkit-animation:load-data-v-58447793 1.2s step-end infinite}@keyframes load-data-v-58447793{0%{width:0}20%{width:5px}40%{width:10px}60%{width:15px}80%{width:20px}}@-webkit-keyframes load-data-v-58447793{0%{width:0}20%{width:5px}40%{width:10px}60%{width:15px}80%{width:20px}}.terminal-last-line[data-v-58447793]{font-size:0;word-spacing:0;letter-spacing:0}.input-box[data-v-58447793]{position:relative;background:#030924;border:none;width:1px;opacity:0;cursor:default}.input-box[data-v-58447793]:focus{outline:none;border:none}',""])},function(t,n){function e(t,n){var e=t[1]||"",r=t[3];if(!r)return e;if(n&&"function"==typeof btoa){var o=i(r);return[e].concat(r.sources.map(function(t){return"/*# sourceURL="+r.sourceRoot+t+" */"})).concat([o]).join("\n")}return[e].join("\n")}function i(t){return"/*# sourceMappingURL=data:application/json;charset=utf-8;base64,"+btoa(unescape(encodeURIComponent(JSON.stringify(t))))+" */"}t.exports=function(t){var n=[];return n.toString=function(){return this.map(function(n){var i=e(n,t);return n[2]?"@media "+n[2]+"{"+i+"}":i}).join("")},n.i=function(t,e){"string"==typeof t&&(t=[[null,t,""]]);for(var i={},r=0;r<this.length;r++){var o=this[r][0];"number"==typeof o&&(i[o]=!0)}for(r=0;r<t.length;r++){var a=t[r];"number"==typeof a[0]&&i[a[0]]||(e&&!a[2]?a[2]=e:e&&(a[2]="("+a[2]+") and ("+e+")"),n.push(a))}},n}},function(t,n,e){function i(t){for(var n=0;n<t.length;n++){var e=t[n],i=l[e.id];if(i){i.refs++;for(var r=0;r<i.parts.length;r++)i.parts[r](e.parts[r]);for(;r<e.parts.length;r++)i.parts.push(o(e.parts[r]));i.parts.length>e.parts.length&&(i.parts.length=e.parts.length)}else{for(var a=[],r=0;r<e.parts.length;r++)a.push(o(e.parts[r]));l[e.id]={id:e.id,refs:1,parts:a}}}}function r(){var t=document.createElement("style");return t.type="text/css",p.appendChild(t),t}function o(t){var n,e,i=document.querySelector("style["+g+'~="'+t.id+'"]');if(i){if(m)return h;i.parentNode.removeChild(i)}if(x){var o=f++;i=d||(d=r()),n=a.bind(null,i,o,!1),e=a.bind(null,i,o,!0)}else i=r(),n=s.bind(null,i),e=function(){i.parentNode.removeChild(i)};return n(t),function(i){if(i){if(i.css===t.css&&i.media===t.media&&i.sourceMap===t.sourceMap)return;n(t=i)}else e()}}function a(t,n,e,i){var r=e?"":i.css;if(t.styleSheet)t.styleSheet.cssText=y(n,r);else{var o=document.createTextNode(r),a=t.childNodes;a[n]&&t.removeChild(a[n]),a.length?t.insertBefore(o,a[n]):t.appendChild(o)}}function s(t,n){var e=n.css,i=n.media,r=n.sourceMap;if(i&&t.setAttribute("media",i),v.ssrId&&t.setAttribute(g,n.id),r&&(e+="\n/*# sourceURL="+r.sources[0]+" */",e+="\n/*# sourceMappingURL=data:application/json;base64,"+btoa(unescape(encodeURIComponent(JSON.stringify(r))))+" */"),t.styleSheet)t.styleSheet.cssText=e;else{for(;t.firstChild;)t.removeChild(t.firstChild);t.appendChild(document.createTextNode(e))}}var u="undefined"!=typeof document;if("undefined"!=typeof DEBUG&&DEBUG&&!u)throw new Error("vue-style-loader cannot be used in a non-browser environment. Use { target: 'node' } in your Webpack config to indicate a server-rendering environment.");var c=e(16),l={},p=u&&(document.head||document.getElementsByTagName("head")[0]),d=null,f=0,m=!1,h=function(){},v=null,g="data-vue-ssr-id",x="undefined"!=typeof navigator&&/msie [6-9]\b/.test(navigator.userAgent.toLowerCase());t.exports=function(t,n,e,r){m=e,v=r||{};var o=c(t,n);return i(o),function(n){for(var e=[],r=0;r<o.length;r++){var a=o[r],s=l[a.id];s.refs--,e.push(s)}n?(o=c(t,n),i(o)):o=[];for(var r=0;r<e.length;r++){var s=e[r];if(0===s.refs){for(var u=0;u<s.parts.length;u++)s.parts[u]();delete l[s.id]}}}};var y=function(){var t=[];return function(n,e){return t[n]=e,t.filter(Boolean).join("\n")}}()},function(t,n){t.exports=function(t,n){for(var e=[],i={},r=0;r<n.length;r++){var o=n[r],a=o[0],s=o[1],u=o[2],c=o[3],l={id:t+":"+r,css:s,media:u,sourceMap:c};i[a]?i[a].parts.push(l):e.push(i[a]={id:a,parts:[l]})}return e}},function(t,n){t.exports=function(t,n,e,i,r,o){var a,s=t=t||{},u=typeof t.default;"object"!==u&&"function"!==u||(a=t,s=t.default);var c="function"==typeof s?s.options:s;n&&(c.render=n.render,c.staticRenderFns=n.staticRenderFns,c._compiled=!0),e&&(c.functional=!0),r&&(c._scopeId=r);var l;if(o?(l=function(t){t=t||this.$vnode&&this.$vnode.ssrContext||this.parent&&this.parent.$vnode&&this.parent.$vnode.ssrContext,t||"undefined"==typeof __VUE_SSR_CONTEXT__||(t=__VUE_SSR_CONTEXT__),i&&i.call(this,t),t&&t._registeredComponents&&t._registeredComponents.add(o)},c._ssrRegister=l):i&&(l=i),l){var p=c.functional,d=p?c.render:c.beforeCreate;p?(c._injectStyles=l,c.render=function(t,n){return l.call(n),d(t,n)}):c.beforeCreate=d?[].concat(d,l):[l]}return{esModule:a,exports:s,options:c}}},function(t,n,e){t.exports={default:e(19),__esModule:!0}},function(t,n,e){e(20),t.exports=e(0).Object.keys},function(t,n,e){var i=e(21),r=e(22);e(34)("keys",function(){return function(t){return r(i(t))}})},function(t,n,e){var i=e(6);t.exports=function(t){return Object(i(t))}},function(t,n,e){var i=e(23),r=e(33);t.exports=Object.keys||function(t){return i(t,r)}},function(t,n,e){var i=e(7),r=e(8),o=e(26)(!1),a=e(29)("IE_PROTO");t.exports=function(t,n){var e,s=r(t),u=0,c=[];for(e in s)e!=a&&i(s,e)&&c.push(e);for(;n.length>u;)i(s,e=n[u++])&&(~o(c,e)||c.push(e));return c}},function(t,n,e){var i=e(25);t.exports=Object("z").propertyIsEnumerable(0)?Object:function(t){return"String"==i(t)?t.split(""):Object(t)}},function(t,n){var e={}.toString;t.exports=function(t){return e.call(t).slice(8,-1)}},function(t,n,e){var i=e(8),r=e(27),o=e(28);t.exports=function(t){return function(n,e,a){var s,u=i(n),c=r(u.length),l=o(a,c);if(t&&e!=e){for(;c>l;)if((s=u[l++])!=s)return!0}else for(;c>l;l++)if((t||l in u)&&u[l]===e)return t||l||0;return!t&&-1}}},function(t,n,e){var i=e(9),r=Math.min;t.exports=function(t){return t>0?r(i(t),9007199254740991):0}},function(t,n,e){var i=e(9),r=Math.max,o=Math.min;t.exports=function(t,n){return t=i(t),t<0?r(t+n,0):o(t,n)}},function(t,n,e){var i=e(30)("keys"),r=e(32);t.exports=function(t){return i[t]||(i[t]=r(t))}},function(t,n,e){var i=e(0),r=e(1),o=r["__core-js_shared__"]||(r["__core-js_shared__"]={});(t.exports=function(t,n){return o[t]||(o[t]=void 0!==n?n:{})})("versions",[]).push({version:i.version,mode:e(31)?"pure":"global",copyright:"© 2019 Denis Pushkarev (zloirock.ru)"})},function(t,n){t.exports=!0},function(t,n){var e=0,i=Math.random();t.exports=function(t){return"Symbol(".concat(void 0===t?"":t,")_",(++e+i).toString(36))}},function(t,n){t.exports="constructor,hasOwnProperty,isPrototypeOf,propertyIsEnumerable,toLocaleString,toString,valueOf".split(",")},function(t,n,e){var i=e(35),r=e(0),o=e(4);t.exports=function(t,n){var e=(r.Object||{})[t]||Object[t],a={};a[t]=n(e),i(i.S+i.F*o(function(){e(1)}),"Object",a)}},function(t,n,e){var i=e(1),r=e(0),o=e(36),a=e(38),s=e(7),u=function(t,n,e){var c,l,p,d=t&u.F,f=t&u.G,m=t&u.S,h=t&u.P,v=t&u.B,g=t&u.W,x=f?r:r[n]||(r[n]={}),y=x.prototype,b=f?i:m?i[n]:(i[n]||{}).prototype;f&&(e=n);for(c in e)(l=!d&&b&&void 0!==b[c])&&s(x,c)||(p=l?b[c]:e[c],x[c]=f&&"function"!=typeof b[c]?e[c]:v&&l?o(p,i):g&&b[c]==p?function(t){var n=function(n,e,i){if(this instanceof t){switch(arguments.length){case 0:return new t;case 1:return new t(n);case 2:return new t(n,e)}return new t(n,e,i)}return t.apply(this,arguments)};return n.prototype=t.prototype,n}(p):h&&"function"==typeof p?o(Function.call,p):p,h&&((x.virtual||(x.virtual={}))[c]=p,t&u.R&&y&&!y[c]&&a(y,c,p)))};u.F=1,u.G=2,u.S=4,u.P=8,u.B=16,u.W=32,u.U=64,u.R=128,t.exports=u},function(t,n,e){var i=e(37);t.exports=function(t,n,e){if(i(t),void 0===n)return t;switch(e){case 1:return function(e){return t.call(n,e)};case 2:return function(e,i){return t.call(n,e,i)};case 3:return function(e,i,r){return t.call(n,e,i,r)}}return function(){return t.apply(n,arguments)}}},function(t,n){t.exports=function(t){if("function"!=typeof t)throw TypeError(t+" is not a function!");return t}},function(t,n,e){var i=e(39),r=e(44);t.exports=e(3)?function(t,n,e){return i.f(t,n,r(1,e))}:function(t,n,e){return t[n]=e,t}},function(t,n,e){var i=e(40),r=e(41),o=e(43),a=Object.defineProperty;n.f=e(3)?Object.defineProperty:function(t,n,e){if(i(t),n=o(n,!0),i(e),r)try{return a(t,n,e)}catch(t){}if("get"in e||"set"in e)throw TypeError("Accessors not supported!");return"value"in e&&(t[n]=e.value),t}},function(t,n,e){var i=e(2);t.exports=function(t){if(!i(t))throw TypeError(t+" is not an object!");return t}},function(t,n,e){t.exports=!e(3)&&!e(4)(function(){return 7!=Object.defineProperty(e(42)("div"),"a",{get:function(){return 7}}).a})},function(t,n,e){var i=e(2),r=e(1).document,o=i(r)&&i(r.createElement);t.exports=function(t){return o?r.createElement(t):{}}},function(t,n,e){var i=e(2);t.exports=function(t,n){if(!i(t))return t;var e,r;if(n&&"function"==typeof(e=t.toString)&&!i(r=e.call(t)))return r;if("function"==typeof(e=t.valueOf)&&!i(r=e.call(t)))return r;if(!n&&"function"==typeof(e=t.toString)&&!i(r=e.call(t)))return r;throw TypeError("Can't convert object to primitive value")}},function(t,n){t.exports=function(t,n){return{enumerable:!(1&t),configurable:!(2&t),writable:!(4&t),value:n}}},function(t,n,e){"use strict";var i=function(){var t=this,n=t.$createElement,e=t._self._c||n;return e("div",{staticClass:"terminal",on:{click:t.handleFocus}},[e("div",{staticStyle:{position:"relative"}},[e("div",{staticClass:"header"},[e("h4",[t._v(t._s(t.title))]),t._v(" "),t._m(0)]),t._v(" "),e("div",{ref:"terminalWindow",staticStyle:{position:"absolute",top:"0",left:"0",right:"0",overflow:"auto","z-index":"1","margin-top":"30px","max-height":"500px"}},[e("div",{staticClass:"terminal-window",attrs:{id:"terminalWindow"}},[e("p",[t._v("Welcome to "+t._s(t.title)+".")]),t._v(" "),e("p",[e("span",{staticClass:"prompt"}),e("span",{staticClass:"cmd"},[t._v("cd "+t._s(t.title))])]),t._v(" "),t._l(t.messageList,function(n,i){return e("p",{key:i},[e("span",[t._v(t._s(n.time))]),t._v(" "),n.label?e("span",{class:n.type},[t._v(t._s(n.label))]):t._e(),t._v(" "),n.message.list?e("span",{staticClass:"cmd"},[e("span",[t._v(t._s(n.message.text))]),t._v(" "),e("ul",t._l(n.message.list,function(n,i){return e("li",{key:i},[n.label?e("span",{class:n.type},[t._v(t._s(n.label)+":")]):t._e(),t._v(" "),e("pre",[t._v(t._s(n.message))])])}),0)]):e("span",{staticClass:"cmd"},[t._v(t._s(n.message))])])}),t._v(" "),t.actionResult?e("p",[e("span",{staticClass:"cmd"},[t._v(t._s(t.actionResult))])]):t._e(),t._v(" "),e("p",{ref:"terminalLastLine",staticClass:"terminal-last-line"},["&nbsp"===t.lastLineContent?e("span",{staticClass:"prompt"},[t._v(" \\"+t._s(t.title)+" ")]):t._e(),t._v(" "),e("span",[t._v(t._s(t.inputCommand))]),t._v(" "),e("span",{class:t.lastLineClass,domProps:{innerHTML:t._s(t.lastLineContent)}}),t._v(" "),e("input",{directives:[{name:"model",rawName:"v-model",value:t.inputCommand,expression:"inputCommand"}],ref:"inputBox",staticClass:"input-box",attrs:{disabled:"&nbsp"!==t.lastLineContent,autofocus:"true",type:"text"},domProps:{value:t.inputCommand},on:{keyup:function(n){t.handleCommand(n)},input:function(n){n.target.composing||(t.inputCommand=n.target.value)}}})])],2)])])])},r=[function(){var t=this,n=t.$createElement,e=t._self._c||n;return e("ul",{staticClass:"shell-dots"},[e("li",{staticClass:"red"}),t._v(" "),e("li",{staticClass:"yellow"}),t._v(" "),e("li",{staticClass:"green"})])}],o={render:i,staticRenderFns:r};n.a=o}])});
//# sourceMappingURL=vue-terminal.min.js.map