var espree = require("espree");
// awesome-custom-parser.js
exports.parseForESLint = function(code, options) {
//   const regex = /{{(?<startdelimiter>-|<|%|\/\*)?\s*(?<statement>(?<keyword>if|range|block|with|define|end|else|prettier-ignore-start|prettier-ignore-end)?[\s\S]*?)\s*(?<endDelimiter>-|>|%|\*\/)?}}|(?<unformattableScript><(script)((?!<)[\s\S])*>((?!<\/script)[\s\S])*?{{[\s\S]*?<\/(script)>)|(?<unformattableStyle><(style)((?!<)[\s\S])*>((?!<\/style)[\s\S])*?{{[\s\S]*?<\/(style)>)/g;
  
  // replace handlebars for eslint
  const regex = /{{.*}}/g
  code = code.replace(regex, '')

    return {
        ast: espree.parse(code, options),
        services: null,
        scopeManager: null,
        visitorKeys: null
    };
};