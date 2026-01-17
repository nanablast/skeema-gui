<template>
  <div class="app">
    <header class="header">
      <h1>🔄 Skeema GUI</h1>
      <span class="subtitle">MySQL Schema & Data Sync Tool</span>
    </header>

    <main class="main">
      <!-- Connection Panel -->
      <div class="connections">
        <ConnectionForm
          title="Source Database"
          :config="sourceConfig"
          :databases="sourceDatabases"
          :loading="sourceLoading"
          :connected="sourceConnected"
          @update:config="sourceConfig = $event"
          @test="testSourceConnection"
          @load-databases="loadSourceDatabases"
        />

        <div class="arrow">➜</div>

        <ConnectionForm
          title="Target Database"
          :config="targetConfig"
          :databases="targetDatabases"
          :loading="targetLoading"
          :connected="targetConnected"
          @update:config="targetConfig = $event"
          @test="testTargetConnection"
          @load-databases="loadTargetDatabases"
        />
      </div>

      <!-- Tab Navigation -->
      <div class="tabs">
        <button
          class="tab"
          :class="{ active: activeTab === 'schema' }"
          @click="activeTab = 'schema'"
        >
          📋 Schema Compare
        </button>
        <button
          class="tab"
          :class="{ active: activeTab === 'data' }"
          @click="activeTab = 'data'"
        >
          📊 Data Sync
        </button>
        <button
          class="tab"
          :class="{ active: activeTab === 'designer' }"
          @click="activeTab = 'designer'"
        >
          🛠️ Table Designer
        </button>
        <button
          class="tab"
          :class="{ active: activeTab === 'browser' }"
          @click="activeTab = 'browser'"
        >
          🗂️ Table Browser
        </button>
      </div>

      <!-- Schema Tab -->
      <div v-show="activeTab === 'schema'">
        <!-- Actions -->
        <div class="actions">
          <button
            class="btn btn-primary"
            @click="compareSchemas"
            :disabled="!canCompare || comparing"
          >
            {{ comparing ? 'Comparing...' : '🔍 Compare Schemas' }}
          </button>
        </div>

        <!-- Results -->
        <DiffResults
          v-if="diffResults.length > 0"
          :results="diffResults"
          :target-config="targetConfig"
          @execute="executeSQL"
        />

        <!-- Empty State -->
        <div v-else-if="hasCompared" class="empty-state">
          ✅ No differences found. Schemas are identical.
        </div>
      </div>

      <!-- Data Sync Tab -->
      <div v-show="activeTab === 'data'">
        <DataSync
          :source-config="sourceConfig"
          :target-config="targetConfig"
          :source-connected="sourceConnected"
          :target-connected="targetConnected"
        />
      </div>

      <!-- Table Designer Tab -->
      <div v-show="activeTab === 'designer'">
        <TableDesigner
          :target-config="targetConfig"
          :target-connected="targetConnected"
        />
      </div>

      <!-- Table Browser Tab -->
      <div v-show="activeTab === 'browser'">
        <TableBrowser
          :config="browserTarget === 'source' ? sourceConfig : targetConfig"
          :connected="browserTarget === 'source' ? sourceConnected : targetConnected"
          :browser-target="browserTarget"
          @switch-target="browserTarget = $event"
        />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import ConnectionForm from './components/ConnectionForm.vue'
import DiffResults from './components/DiffResults.vue'
import DataSync from './components/DataSync.vue'
import TableDesigner from './components/TableDesigner.vue'
import TableBrowser from './components/TableBrowser.vue'
import { TestConnection, GetDatabases, CompareSchemas, ExecuteSQL } from '../wailsjs/go/main/App'
import { database } from '../wailsjs/go/models'

type ConnectionConfig = database.ConnectionConfig
type DiffResult = database.DiffResult

// Active tab
const activeTab = ref<'schema' | 'data' | 'designer' | 'browser'>('schema')

// Browser target switch
const browserTarget = ref<'source' | 'target'>('target')

// Source connection
const sourceConfig = ref<ConnectionConfig>({
  host: 'localhost',
  port: 3306,
  user: 'root',
  password: '',
  database: ''
})
const sourceDatabases = ref<string[]>([])
const sourceLoading = ref(false)
const sourceConnected = ref(false)

// Target connection
const targetConfig = ref<ConnectionConfig>({
  host: 'localhost',
  port: 3306,
  user: 'root',
  password: '',
  database: ''
})
const targetDatabases = ref<string[]>([])
const targetLoading = ref(false)
const targetConnected = ref(false)

// Comparison state
const comparing = ref(false)
const hasCompared = ref(false)
const diffResults = ref<DiffResult[]>([])

const canCompare = computed(() => {
  return sourceConnected.value &&
         targetConnected.value &&
         sourceConfig.value.database &&
         targetConfig.value.database
})

async function testSourceConnection() {
  sourceLoading.value = true
  try {
    await TestConnection(sourceConfig.value)
    sourceConnected.value = true
    await loadSourceDatabases()
  } catch (e: any) {
    alert('Connection failed: ' + e)
    sourceConnected.value = false
  } finally {
    sourceLoading.value = false
  }
}

async function loadSourceDatabases() {
  try {
    sourceDatabases.value = await GetDatabases(sourceConfig.value)
  } catch (e: any) {
    console.error(e)
  }
}

async function testTargetConnection() {
  targetLoading.value = true
  try {
    await TestConnection(targetConfig.value)
    targetConnected.value = true
    await loadTargetDatabases()
  } catch (e: any) {
    alert('Connection failed: ' + e)
    targetConnected.value = false
  } finally {
    targetLoading.value = false
  }
}

async function loadTargetDatabases() {
  try {
    targetDatabases.value = await GetDatabases(targetConfig.value)
  } catch (e: any) {
    console.error(e)
  }
}

async function compareSchemas() {
  comparing.value = true
  hasCompared.value = false
  diffResults.value = []

  try {
    const results = await CompareSchemas(sourceConfig.value, targetConfig.value)
    diffResults.value = results || []
    hasCompared.value = true
  } catch (e: any) {
    alert('Comparison failed: ' + e)
  } finally {
    comparing.value = false
  }
}

async function executeSQL(sql: string) {
  try {
    await ExecuteSQL(targetConfig.value, sql)
    alert('SQL executed successfully!')
    // Re-compare after execution
    await compareSchemas()
  } catch (e: any) {
    alert('Execution failed: ' + e)
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
  background: #1a1a2e;
  color: #eee;
  min-height: 100vh;
}

.app {
  min-height: 100vh;
  padding: 20px;
}

.header {
  text-align: center;
  margin-bottom: 30px;
}

.header h1 {
  font-size: 28px;
  color: #4fc3f7;
  margin-bottom: 5px;
}

.subtitle {
  color: #888;
  font-size: 14px;
}

.connections {
  display: flex;
  gap: 20px;
  align-items: flex-start;
  justify-content: center;
  margin-bottom: 20px;
}

.arrow {
  font-size: 32px;
  color: #4fc3f7;
  padding-top: 80px;
}

.actions {
  text-align: center;
  margin: 20px 0;
}

.btn {
  padding: 12px 30px;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #4fc3f7;
  color: #1a1a2e;
  font-weight: bold;
}

.btn-primary:hover:not(:disabled) {
  background: #29b6f6;
  transform: translateY(-1px);
}

.btn-primary:disabled {
  background: #555;
  color: #888;
  cursor: not-allowed;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: #4caf50;
  font-size: 18px;
}

.tabs {
  display: flex;
  gap: 10px;
  justify-content: center;
  margin-bottom: 20px;
}

.tab {
  padding: 10px 24px;
  border: none;
  border-radius: 6px;
  background: #0f3460;
  color: #888;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.tab:hover {
  color: #4fc3f7;
}

.tab.active {
  background: #4fc3f7;
  color: #1a1a2e;
  font-weight: bold;
}
</style>
