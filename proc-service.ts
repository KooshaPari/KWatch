#!/usr/bin/env node

import * as blessed from 'blessed';
import * as chokidar from 'chokidar';
import { exec } from 'child_process';
import { promisify } from 'util';
import * as fs from 'fs/promises';
import * as path from 'path';
import * as http from 'http';
import * as url from 'url';

const execAsync = promisify(exec);

interface CommandResult {
  command: string;
  passed: boolean;
  issueCount: number;
  output: string;
  timestamp: Date;
  duration: number;
}

interface RunResult {
  timestamp: Date;
  tsc: CommandResult;
  lint: CommandResult;
  test: CommandResult;
  build: CommandResult;
  duration: number;
}

class ProcService {
  private screen!: blessed.Widgets.Screen;
  private table!: blessed.Widgets.TableElement;
  private logBox!: blessed.Widgets.BoxElement;
  private statusBox!: blessed.Widgets.BoxElement;
  private history: RunResult[] = [];
  private isRunning = false;
  private watcher!: chokidar.FSWatcher;
  private server!: http.Server;
  private lastFileChange = 0;
  private debounceMs = 2000; // 2 second debounce for file changes

  constructor() {
    this.setupTUI();
    this.setupFileWatcher();
    this.setupHTTPServer();
    this.startPeriodicCheck();
    this.loadHistory();
  }

  private setupTUI() {
    this.screen = blessed.screen({
      smartCSR: true,
      title: 'Proc Service Monitor'
    });

    // Status box at the top
    this.statusBox = blessed.box({
      top: 0,
      left: 0,
      width: '100%',
      height: 3,
      content: 'Process Service - Starting...',
      tags: true,
      border: {
        type: 'line'
      },
      style: {
        fg: 'white',
        bg: 'blue',
        border: {
          fg: '#f0f0f0'
        }
      }
    });

    // History table
    this.table = blessed.table({
      top: 3,
      left: 0,
      width: '100%',
      height: '70%',
      border: {
        type: 'line'
      },
      style: {
        fg: 'white',
        border: {
          fg: '#f0f0f0'
        },
        header: {
          fg: 'blue',
          bold: true
        }
      },
      columnSpacing: 2,
      columnWidth: [20, 8, 8, 8, 8, 8]
    });

    // Log box at the bottom
    this.logBox = blessed.box({
      top: '73%',
      left: 0,
      width: '100%',
      height: '27%',
      border: {
        type: 'line'
      },
      style: {
        fg: 'white',
        border: {
          fg: '#f0f0f0'
        }
      },
      scrollable: true,
      alwaysScroll: true,
      mouse: true,
      keys: true,
      tags: true
    });

    this.screen.append(this.statusBox);
    this.screen.append(this.table);
    this.screen.append(this.logBox);

    // Quit on Escape, q, or Control-C
    this.screen.key(['escape', 'q', 'C-c'], () => {
      this.cleanup();
      process.exit(0);
    });

    this.screen.render();
    this.updateTable();
  }

  private setupFileWatcher() {
    this.watcher = chokidar.watch([
      'src/**/*',
      'package.json',
      'tsconfig.json',
      '*.config.js',
      '*.config.ts',
      'tests/**/*',
      '__tests__/**/*'
    ], {
      ignored: [
        'node_modules/**/*',
        'dist/**/*',
        'build/**/*',
        '.git/**/*',
        '**/*.log'
      ],
      persistent: true
    });

    this.watcher.on('change', (path) => {
      this.lastFileChange = Date.now();
      this.log(`File changed: ${path}`);
    });

    this.watcher.on('add', (path) => {
      this.lastFileChange = Date.now();
      this.log(`File added: ${path}`);
    });

    this.watcher.on('unlink', (path) => {
      this.lastFileChange = Date.now();
      this.log(`File removed: ${path}`);
    });
  }

  private setupHTTPServer() {
    this.server = http.createServer((req, res) => {
      const parsedUrl = url.parse(req.url!, true);
      
      res.setHeader('Content-Type', 'application/json');
      res.setHeader('Access-Control-Allow-Origin', '*');

      switch (parsedUrl.pathname) {
        case '/status':
          res.end(JSON.stringify({
            isRunning: this.isRunning,
            lastRun: this.history[0] || null,
            history: this.history.slice(0, 10)
          }, null, 2));
          break;

        case '/history':
          res.end(JSON.stringify(this.history, null, 2));
          break;

        case '/run':
          if (!this.isRunning) {
            this.runChecks().then(() => {
              res.end(JSON.stringify({ message: 'Run completed' }));
            }).catch(err => {
              res.statusCode = 500;
              res.end(JSON.stringify({ error: err.message }));
            });
          } else {
            res.statusCode = 429;
            res.end(JSON.stringify({ error: 'Already running' }));
          }
          break;

        default:
          res.statusCode = 404;
          res.end(JSON.stringify({ error: 'Not found' }));
      }
    });

    this.server.listen(3737, () => {
      this.log('HTTP server listening on port 3737');
      this.log('CLI usage: curl http://localhost:3737/status');
    });
  }

  private async startPeriodicCheck() {
    setInterval(async () => {
      if (this.lastFileChange > 0 && 
          Date.now() - this.lastFileChange > this.debounceMs && 
          !this.isRunning) {
        this.lastFileChange = 0;
        this.log('File changes detected, running checks...');
        await this.runChecks();
      }
    }, 5000); // Check every 5 seconds

    // Also run every minute regardless
    setInterval(async () => {
      if (!this.isRunning) {
        this.log('Periodic check...');
        await this.runChecks();
      }
    }, 60000); // Every minute
  }

  private async runChecks(): Promise<void> {
    if (this.isRunning) {
      this.log('Already running checks, skipping...');
      return;
    }

    this.isRunning = true;
    this.updateStatus('Running checks...');
    
    const startTime = Date.now();
    
    try {
      const commands = [
        { name: 'tsc', cmd: 'npx tsc --noEmit 2>&1' },
        { name: 'lint', cmd: 'bun run lint' },
        { name: 'test', cmd: 'bun run test' },
        { name: 'build', cmd: 'bun run build' }
      ];

      // Run all commands in parallel
      const results = await Promise.allSettled(
        commands.map(({ name, cmd }) => this.runCommand(name, cmd))
      );

      const runResult: RunResult = {
        timestamp: new Date(),
        tsc: results[0].status === 'fulfilled' ? results[0].value : this.createErrorResult('tsc', 'Failed to run'),
        lint: results[1].status === 'fulfilled' ? results[1].value : this.createErrorResult('lint', 'Failed to run'),
        test: results[2].status === 'fulfilled' ? results[2].value : this.createErrorResult('test', 'Failed to run'),
        build: results[3].status === 'fulfilled' ? results[3].value : this.createErrorResult('build', 'Failed to run'),
        duration: Date.now() - startTime
      };

      this.history.unshift(runResult);
      if (this.history.length > 50) {
        this.history = this.history.slice(0, 50);
      }

      this.updateTable();
      this.saveHistory();
      
      const passed = [runResult.tsc, runResult.lint, runResult.test, runResult.build]
        .filter(r => r.passed).length;
      
      this.updateStatus(`Completed: ${passed}/4 passed`);
      this.log(`Run completed in ${runResult.duration}ms - ${passed}/4 passed`);
    } catch (error) {
      this.log(`Error running checks: ${error}`);
      this.updateStatus('Error running checks');
    } finally {
      this.isRunning = false;
    }
  }

  private async runCommand(name: string, cmd: string): Promise<CommandResult> {
    const startTime = Date.now();
    
    try {
      const { stdout, stderr } = await execAsync(cmd, { 
        timeout: 300000, // 5 minute timeout
        maxBuffer: 10 * 1024 * 1024 // 10MB buffer
      });
      
      const output = (stdout + stderr).trim();
      const issueCount = this.countIssues(output);
      
      return {
        command: cmd,
        passed: true,
        issueCount,
        output,
        timestamp: new Date(),
        duration: Date.now() - startTime
      };
    } catch (error: any) {
      const output = (error.stdout || '') + (error.stderr || '') + error.message;
      const issueCount = this.countIssues(output);
      
      return {
        command: cmd,
        passed: false,
        issueCount,
        output,
        timestamp: new Date(),
        duration: Date.now() - startTime
      };
    }
  }

  private createErrorResult(command: string, message: string): CommandResult {
    return {
      command,
      passed: false,
      issueCount: 1,
      output: message,
      timestamp: new Date(),
      duration: 0
    };
  }

  private countIssues(output: string): number {
    const lines = output.split('\n');
    return lines.filter(line => 
      /\b(error|warning|fail|failed)\b/i.test(line)
    ).length;
  }

  private updateTable() {
    const headers = ['Timestamp', 'TSC', 'Lint', 'Test', 'Build', 'Duration'];
    const data = [headers];

    this.history.forEach(run => {
      const formatResult = (result: CommandResult) => {
        const symbol = result.passed ? '✓' : '✗';
        const count = result.issueCount > 0 ? ` (${result.issueCount})` : '';
        return `${symbol}${count}`;
      };

      data.push([
        run.timestamp.toLocaleTimeString(),
        formatResult(run.tsc),
        formatResult(run.lint),
        formatResult(run.test),
        formatResult(run.build),
        `${run.duration}ms`
      ]);
    });

    this.table.setData(data);
    this.screen.render();
  }

  private updateStatus(status: string) {
    const timestamp = new Date().toLocaleTimeString();
    this.statusBox.setContent(`{bold}Proc Service{/bold} - ${status} - ${timestamp} - Server: :3737`);
    this.screen.render();
  }

  private log(message: string) {
    const timestamp = new Date().toLocaleTimeString();
    const content = this.logBox.getContent();
    const newContent = content + `[${timestamp}] ${message}\n`;
    this.logBox.setContent(newContent);
    this.logBox.scrollTo(this.logBox.getScrollHeight());
    this.screen.render();
  }

  private async saveHistory() {
    try {
      await fs.writeFile('.proc-history.json', JSON.stringify(this.history, null, 2));
    } catch (error) {
      this.log(`Failed to save history: ${error}`);
    }
  }

  private async loadHistory() {
    try {
      const data = await fs.readFile('.proc-history.json', 'utf-8');
      this.history = JSON.parse(data);
      this.history.forEach(run => {
        run.timestamp = new Date(run.timestamp);
        run.tsc.timestamp = new Date(run.tsc.timestamp);
        run.lint.timestamp = new Date(run.lint.timestamp);
        run.test.timestamp = new Date(run.test.timestamp);
        run.build.timestamp = new Date(run.build.timestamp);
      });
      this.updateTable();
      this.log(`Loaded ${this.history.length} history entries`);
    } catch (error) {
      this.log('No previous history found');
    }
  }

  private cleanup() {
    if (this.watcher) {
      this.watcher.close();
    }
    if (this.server) {
      this.server.close();
    }
  }

  public start() {
    this.log('Proc Service started');
    this.updateStatus('Ready - Watching for changes');
    
    // Run initial check
    setTimeout(() => {
      this.runChecks();
    }, 2000);
  }
}

// Start the service
const service = new ProcService();
service.start();
