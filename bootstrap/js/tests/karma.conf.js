/* eslint-env node */
const path = require('path')
const ip = require('ip')
const babel = require('rollup-plugin-babel')
const istanbul = require('rollup-plugin-istanbul')
const resolve = require('rollup-plugin-node-resolve')

const {
  browsers,
  browsersKeys
} = require('./browsers')

const { env } = process
const browserStack = env.BROWSER === 'true'
const debug = env.DEBUG === 'true'
const frameworks = [
  'jasmine'
]

const plugins = [
  'karma-jasmine',
  'karma-rollup-preprocessor'
]

const reporters = ['dots']

const detectBrowsers = {
  usePhantomJS: false,
  postDetection(availableBrowser) {
    if (env.CI === true || availableBrowser.includes('Chrome')) {
      return debug ? ['Chrome'] : ['ChromeHeadless']
    }

    if (availableBrowser.includes('Firefox')) {
      return debug ? ['Firefox'] : ['FirefoxHeadless']
    }

    throw new Error('Please install Firefox or Chrome')
  }
}

const customLaunchers = {
  FirefoxHeadless: {
    base: 'Firefox',
    flags: ['-headless']
  }
}

const conf = {
  basePath: '../..',
  port: 9876,
  colors: true,
  autoWatch: false,
  singleRun: true,
  concurrency: Infinity,
  client: {
    clearContext: false
  },
  files: [
    'node_modules/hammer-simulator/index.js',
    { pattern: 'js/tests/unit/**/*.spec.js', watched: !browserStack }
  ],
  preprocessors: {
    'js/tests/unit/**/*.spec.js': ['rollup']
  },
  rollupPreprocessor: {
    plugins: [
      istanbul({
        exclude: ['js/tests/unit/**/*.spec.js', 'js/tests/helpers/**/*.js']
      }),
      babel({
        // Only transpile our source code
        exclude: 'node_modules/**',
        // Include only required helpers
        externalHelpersWhitelist: [
          'defineProperties',
          'createClass',
          'inheritsLoose',
          'defineProperty',
          'objectSpread2'
        ],
        plugins: [
          '@babel/plugin-proposal-object-rest-spread'
        ]
      }),
      resolve()
    ],
    output: {
      format: 'iife',
      name: 'bootstrapTest',
      sourcemap: 'inline'
    }
  }
}

if (browserStack) {
  conf.hostname = ip.address()
  conf.browserStack = {
    username: env.BROWSER_STACK_USERNAME,
    accessKey: env.BROWSER_STACK_ACCESS_KEY,
    build: `bootstrap-${new Date().toISOString()}`,
    project: 'Bootstrap',
    retryLimit: 2
  }
  plugins.push('karma-browserstack-launcher', 'karma-jasmine-html-reporter')
  conf.customLaunchers = browsers
  conf.browsers = browsersKeys
  reporters.push('BrowserStack', 'kjhtml')
} else {
  frameworks.push('detectBrowsers')
  plugins.push(
    'karma-chrome-launcher',
    'karma-firefox-launcher',
    'karma-detect-browsers',
    'karma-coverage-istanbul-reporter'
  )
  reporters.push('coverage-istanbul')
  conf.customLaunchers = customLaunchers
  conf.detectBrowsers = detectBrowsers
  conf.coverageIstanbulReporter = {
    dir: path.resolve(__dirname, '../coverage/'),
    reports: ['lcov', 'text-summary'],
    thresholds: {
      emitWarning: false,
      global: {
        statements: 90,
        branches: 90,
        functions: 90,
        lines: 90
      },
      each: {
        overrides: {
          'js/src/dom/polyfill.js': {
            statements: 39,
            lines: 37,
            branches: 19,
            functions: 50
          }
        }
      }
    }
  }

  if (debug) {
    conf.hostname = ip.address()
    plugins.push('karma-jasmine-html-reporter')
    reporters.push('kjhtml')
    conf.singleRun = false
    conf.autoWatch = true
  }
}

conf.frameworks = frameworks
conf.plugins = plugins
conf.reporters = reporters

module.exports = karmaConfig => {
  // possible values: karmaConfig.LOG_DISABLE || karmaConfig.LOG_ERROR || karmaConfig.LOG_WARN || karmaConfig.LOG_INFO || karmaConfig.LOG_DEBUG
  conf.logLevel = karmaConfig.LOG_ERROR || karmaConfig.LOG_WARN
  karmaConfig.set(conf)
}
