#!/usr/bin/env node

import * as http from 'http';

class ProcCLI {
  private baseUrl = 'http://localhost:3737';

  async request(path: string): Promise<any> {
    return new Promise((resolve, reject) => {
      const req = http.get(`${this.baseUrl}${path}`, (res) => {
        let data = '';
        
        res.on('data', (chunk) => {
          data += chunk;
        });
        
        res.on('end', () => {
          try {
            const json = JSON.parse(data);
            resolve(json);
          } catch (error) {
            reject(new Error(`Failed to parse JSON: ${error}`));
          }
        });
      });
      
      req.on('error', (error) => {
        reject(new Error(`Request failed: ${error.message}`));
      });
      
      req.setTimeout(5000, () => {
        req.destroy();
        reject(new Error('Request timeout'));
      });
    });
  }

  async status() {
    try {
      const data = await this.request('/status');
      
      console.log('\n📊 Process Service Status');
      console.log('═'.repeat(50));
      console.log(`Running: ${data.isRunning ? '🔄 Yes' : '⏸️  No'}`);
      
      if (data.lastRun) {
        const run = data.lastRun;
        const timestamp = new Date(run.timestamp).toLocaleString();
        
        console.log(`\n🕒 Last Run: ${timestamp} (${run.duration}ms)`);
        console.log('─'.repeat(50));
        
        const results = [
          ['TSC', run.tsc],
          ['Lint', run.lint],
          ['Test', run.test],
          ['Build', run.build]
        ];
        
        results.forEach(([name, result]) => {
          const status = result.passed ? '✅' : '❌';
          const issues = result.issueCount > 0 ? ` (${result.issueCount} issues)` : '';
          console.log(`${status} ${name.padEnd(6)} ${result.duration}ms${issues}`);
        });
        
        const passed = results.filter(([_, r]) => r.passed).length;
        console.log(`\n📈 Overall: ${passed}/4 passed`);
      } else {
        console.log('\n⚠️  No runs yet');
      }
      
      return data;
    } catch (error) {
      console.error(`❌ Error: ${error instanceof Error ? error.message : String(error)}`);
      console.error('💡 Make sure the proc service is running with: npm run proc:start');
      process.exit(1);
    }
  }

  async history(limit = 10) {
    try {
      const data = await this.request('/history');
      
      console.log('\n📜 Process Service History');
      console.log('═'.repeat(80));
      
      if (data.length === 0) {
        console.log('⚠️  No history available');
        return;
      }
      
      const runs = data.slice(0, limit);
      
      // Header
      console.log('Time'.padEnd(12) + 'TSC'.padEnd(8) + 'Lint'.padEnd(8) + 'Test'.padEnd(8) + 'Build'.padEnd(8) + 'Duration');
      console.log('─'.repeat(80));
      
      runs.forEach((run: any) => {
        const time = new Date(run.timestamp).toLocaleTimeString();
        const formatResult = (result: any) => {
          const symbol = result.passed ? '✓' : '✗';
          const count = result.issueCount > 0 ? result.issueCount : '';
          return `${symbol}${count}`.padEnd(7);
        };
        
        console.log(
          time.padEnd(12) +
          formatResult(run.tsc) +
          formatResult(run.lint) +
          formatResult(run.test) +
          formatResult(run.build) +
          `${run.duration}ms`
        );
      });
      
      return data;
    } catch (error) {
      console.error(`❌ Error: ${error instanceof Error ? error.message : String(error)}`);
      process.exit(1);
    }
  }

  async run() {
    try {
      console.log('🚀 Triggering manual run...');
      const data = await this.request('/run');
      console.log('✅ Run completed successfully');
      
      // Show status after run
      setTimeout(() => this.status(), 1000);
      
      return data;
    } catch (error) {
      console.error(`❌ Error: ${error instanceof Error ? error.message : String(error)}`);
      process.exit(1);
    }
  }

  async watch() {
    console.log('👀 Watching process service status...');
    console.log('Press Ctrl+C to stop\n');
    
    const showStatus = async () => {
      try {
        process.stdout.write('\x1B[2J\x1B[0f'); // Clear screen
        await this.status();
      } catch (error) {
        console.error(`Error: ${error instanceof Error ? error.message : String(error)}`);
      }
    };
    
    // Initial status
    await showStatus();
    
    // Update every 5 seconds
    const interval = setInterval(showStatus, 5000);
    
    process.on('SIGINT', () => {
      clearInterval(interval);
      console.log('\n👋 Stopped watching');
      process.exit(0);
    });
  }

  showHelp() {
    console.log(`
🔧 Proc Service CLI

Usage: proc-cli <command>

Commands:
  status     Show current status and last run results
  history    Show run history (default: last 10)
  run        Trigger a manual run
  watch      Watch status in real-time
  help       Show this help

Examples:
  proc-cli status
  proc-cli history
  proc-cli run
  proc-cli watch

The proc service must be running for these commands to work.
Start it with: npm run proc:start
`);
  }
}

// CLI interface
const cli = new ProcCLI();
const command = process.argv[2];

switch (command) {
  case 'status':
    cli.status();
    break;
  
  case 'history':
    const limit = parseInt(process.argv[3]) || 10;
    cli.history(limit);
    break;
  
  case 'run':
    cli.run();
    break;
  
  case 'watch':
    cli.watch();
    break;
  
  case 'help':
  case '--help':
  case '-h':
    cli.showHelp();
    break;
  
  default:
    if (!command) {
      cli.showHelp();
    } else {
      console.error(`❌ Unknown command: ${command}`);
      cli.showHelp();
      process.exit(1);
    }
}
