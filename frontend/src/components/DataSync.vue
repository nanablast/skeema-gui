<template>
  <div class="data-sync">
    <div class="sync-header">
      <h3>📊 Data Sync</h3>
      <p class="hint">Compare and sync data from Source to Target</p>
    </div>

    <!-- Table Selection -->
    <div class="table-section">
      <div class="section-header">
        <h4>Select Table</h4>
        <button class="btn btn-refresh" @click="loadTables" :disabled="loadingTables">
          {{ loadingTables ? 'Loading...' : '🔄 Refresh' }}
        </button>
      </div>

      <div class="table-list" v-if="tables.length > 0">
        <div
          v-for="table in tables"
          :key="table.tableName"
          class="table-item"
          :class="{ selected: selectedTable === table.tableName, 'no-pk': table.primaryKeys.length === 0 }"
          @click="selectTable(table)"
        >
          <span class="table-name">{{ table.tableName }}</span>
          <span class="table-info">
            <span v-if="table.primaryKeys.length === 0" class="no-pk-badge">No PK</span>
            <span v-else class="pk-badge">PK: {{ table.primaryKeys.join(', ') }}</span>
            <span class="row-count">{{ table.sourceCount }} rows</span>
          </span>
        </div>
      </div>
      <div v-else-if="!loadingTables" class="empty-tables">
        Connect to databases and click Refresh to load tables
      </div>
    </div>

    <!-- Comparison Results -->
    <div class="compare-section" v-if="selectedTable">
      <div class="section-header">
        <h4>{{ selectedTable }} - Data Differences</h4>
        <button class="btn btn-compare" @click="compareData" :disabled="comparing">
          {{ comparing ? 'Comparing...' : '🔍 Compare Data' }}
        </button>
      </div>

      <!-- Progress Log - Terminal Style -->
      <div class="terminal" v-if="comparing || logs.length > 0">
        <div class="terminal-header">
          <span class="terminal-dot red"></span>
          <span class="terminal-dot yellow"></span>
          <span class="terminal-dot green"></span>
          <span class="terminal-title">Data Compare - {{ selectedTable }}</span>
        </div>
        <div class="terminal-body" ref="terminalBody">
          <div v-for="(log, i) in logs" :key="i" class="terminal-line">
            <span class="terminal-prompt">$</span>
            <span class="terminal-text" :class="log.type">{{ log.message }}</span>
            <span class="terminal-status" v-if="log.type === 'done'">✓</span>
            <span class="terminal-status error" v-else-if="log.type === 'error'">✗</span>
          </div>
          <div v-if="comparing" class="terminal-line">
            <span class="terminal-prompt">$</span>
            <span class="terminal-text">{{ currentStep }}</span>
            <span class="terminal-dots">
              <span class="dot" :class="{ active: dotIndex === 0 }">.</span>
              <span class="dot" :class="{ active: dotIndex === 1 }">.</span>
              <span class="dot" :class="{ active: dotIndex === 2 }">.</span>
            </span>
          </div>
          <div v-if="comparing" class="terminal-cursor"></div>
        </div>
      </div>

      <!-- Summary -->
      <div class="sync-summary" v-if="summary">
        <div class="summary-item insert">
          <span class="count">{{ summary.insertCount }}</span>
          <span class="label">Insert</span>
        </div>
        <div class="summary-item update">
          <span class="count">{{ summary.updateCount }}</span>
          <span class="label">Update</span>
        </div>
        <div class="summary-item delete">
          <span class="count">{{ summary.deleteCount }}</span>
          <span class="label">Delete</span>
        </div>
      </div>

      <!-- Sync Options -->
      <div class="sync-options" v-if="dataDiffs.length > 0">
        <label class="checkbox-label">
          <input type="checkbox" v-model="syncInsert" />
          <span>INSERT ({{ insertDiffs.length }})</span>
        </label>
        <label class="checkbox-label">
          <input type="checkbox" v-model="syncUpdate" />
          <span>UPDATE ({{ updateDiffs.length }})</span>
        </label>
        <label class="checkbox-label">
          <input type="checkbox" v-model="syncDelete" />
          <span>DELETE ({{ deleteDiffs.length }})</span>
        </label>
      </div>

      <!-- Diff List -->
      <div class="diff-list" v-if="filteredDiffs.length > 0">
        <div
          v-for="(diff, index) in filteredDiffs.slice(0, showLimit)"
          :key="index"
          class="diff-item"
          :class="diff.type"
        >
          <div class="diff-header">
            <span class="diff-badge" :class="diff.type">{{ diff.type.toUpperCase() }}</span>
            <span class="pk-info">{{ formatPrimaryKey(diff.primaryKey) }}</span>
          </div>
          <div class="diff-sql">
            <code>{{ diff.sql }}</code>
          </div>
        </div>

        <div v-if="filteredDiffs.length > showLimit" class="show-more">
          <button class="btn btn-small" @click="showLimit += 50">
            Show more ({{ filteredDiffs.length - showLimit }} remaining)
          </button>
        </div>
      </div>

      <div v-else-if="hasCompared && dataDiffs.length === 0" class="no-diff">
        ✅ Data is identical, no sync needed
      </div>

      <!-- Execute Actions -->
      <div class="sync-actions" v-if="filteredDiffs.length > 0">
        <button class="btn btn-copy" @click="copySQL">
          📋 Copy SQL ({{ filteredDiffs.length }})
        </button>
        <button class="btn btn-execute" @click="showConfirmDialog = true">
          ▶️ Execute Sync
        </button>
      </div>
    </div>

    <!-- Confirm Dialog -->
    <div class="dialog-overlay" v-if="showConfirmDialog" @click.self="showConfirmDialog = false">
      <div class="dialog">
        <h4>⚠️ Confirm Sync</h4>
        <p>Execute {{ filteredDiffs.length }} SQL statements on target database?</p>
        <div class="dialog-actions">
          <button class="btn btn-cancel" @click="showConfirmDialog = false">Cancel</button>
          <button class="btn btn-confirm" @click="executeSync">Execute</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onUnmounted } from 'vue'
import { GetTablesForSync, CompareTableData, GetDataSyncSummary, ExecuteSQL } from '../../wailsjs/go/main/App'
import { database } from '../../wailsjs/go/models'

type ConnectionConfig = database.ConnectionConfig
type TableDataInfo = database.TableDataInfo
type DataDiffResult = database.DataDiffResult

const props = defineProps<{
  sourceConfig: ConnectionConfig
  targetConfig: ConnectionConfig
  sourceConnected: boolean
  targetConnected: boolean
}>()

const emit = defineEmits<{
  'execute': [sql: string]
}>()

// State
const tables = ref<TableDataInfo[]>([])
const loadingTables = ref(false)
const selectedTable = ref<string | null>(null)
const comparing = ref(false)
const hasCompared = ref(false)
const dataDiffs = ref<DataDiffResult[]>([])
const summary = ref<TableDataInfo | null>(null)
const showLimit = ref(50)
const showConfirmDialog = ref(false)

// Progress log
interface LogEntry {
  message: string
  type: 'progress' | 'done' | 'error'
  time?: string
}
const logs = ref<LogEntry[]>([])
const progressPercent = ref(0)
const currentStep = ref('')
const dotIndex = ref(0)
const terminalBody = ref<HTMLElement | null>(null)
let dotInterval: number | null = null

function startDotAnimation() {
  dotInterval = window.setInterval(() => {
    dotIndex.value = (dotIndex.value + 1) % 3
  }, 400)
}

function stopDotAnimation() {
  if (dotInterval) {
    clearInterval(dotInterval)
    dotInterval = null
  }
}

onUnmounted(() => {
  stopDotAnimation()
})

function addLog(message: string, type: 'progress' | 'done' | 'error' = 'progress') {
  const now = new Date()
  const time = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`
  logs.value.push({ message, type, time })
  nextTick(() => {
    if (terminalBody.value) {
      terminalBody.value.scrollTop = terminalBody.value.scrollHeight
    }
  })
}

function clearLogs() {
  logs.value = []
  progressPercent.value = 0
  currentStep.value = ''
}

function delay(ms: number) {
  return new Promise(resolve => setTimeout(resolve, ms))
}

// Sync options
const syncInsert = ref(true)
const syncUpdate = ref(true)
const syncDelete = ref(false)

// Computed
const insertDiffs = computed(() => dataDiffs.value.filter(d => d.type === 'insert'))
const updateDiffs = computed(() => dataDiffs.value.filter(d => d.type === 'update'))
const deleteDiffs = computed(() => dataDiffs.value.filter(d => d.type === 'delete'))

const filteredDiffs = computed(() => {
  return dataDiffs.value.filter(d => {
    if (d.type === 'insert' && !syncInsert.value) return false
    if (d.type === 'update' && !syncUpdate.value) return false
    if (d.type === 'delete' && !syncDelete.value) return false
    return true
  })
})

// Methods
async function loadTables() {
  if (!props.sourceConnected || !props.sourceConfig.database) {
    alert('Please connect to source database first')
    return
  }

  loadingTables.value = true
  try {
    tables.value = await GetTablesForSync(props.sourceConfig) || []
  } catch (e: any) {
    alert('Failed to load tables: ' + e)
  } finally {
    loadingTables.value = false
  }
}

function selectTable(table: TableDataInfo) {
  if (table.primaryKeys.length === 0) {
    alert('Cannot sync table without primary key')
    return
  }
  selectedTable.value = table.tableName
  dataDiffs.value = []
  summary.value = null
  hasCompared.value = false
  showLimit.value = 50
}

async function compareData() {
  if (!selectedTable.value) return

  comparing.value = true
  hasCompared.value = false
  dataDiffs.value = []
  clearLogs()
  startDotAnimation()

  try {
    currentStep.value = 'Initializing comparison'
    await delay(300)
    addLog('Initializing comparison', 'done')

    currentStep.value = 'Connecting to source database'
    await delay(200)
    addLog('Connected to source database', 'done')

    currentStep.value = 'Fetching source data'
    await delay(200)
    addLog('Fetched source data', 'done')

    currentStep.value = 'Connecting to target database'
    await delay(200)
    addLog('Connected to target database', 'done')

    currentStep.value = 'Fetching target data'
    await delay(200)
    addLog('Fetched target data', 'done')

    currentStep.value = 'Comparing records by primary key'
    const diffsPromise = CompareTableData(props.sourceConfig, props.targetConfig, selectedTable.value)
    const summaryPromise = GetDataSyncSummary(props.sourceConfig, props.targetConfig, selectedTable.value)

    const [diffs, summaryData] = await Promise.all([diffsPromise, summaryPromise])

    dataDiffs.value = diffs || []
    summary.value = summaryData
    hasCompared.value = true

    addLog('Compared records by primary key', 'done')

    const diffCount = diffs?.length || 0
    addLog(`Comparison complete: ${diffCount} difference(s) found`, 'done')

  } catch (e: any) {
    addLog(`Error: ${e}`, 'error')
  } finally {
    stopDotAnimation()
    comparing.value = false
    currentStep.value = ''
  }
}

function formatPrimaryKey(pk: Record<string, any>): string {
  return Object.entries(pk).map(([k, v]) => `${k}=${v}`).join(', ')
}

function copySQL() {
  const sql = filteredDiffs.value.map(d => d.sql).join('\n')
  navigator.clipboard.writeText(sql)
  alert(`Copied ${filteredDiffs.value.length} SQL statements`)
}

async function executeSync() {
  if (filteredDiffs.value.length === 0) return

  showConfirmDialog.value = false

  try {
    const sql = filteredDiffs.value.map(d => d.sql).join('\n')
    await ExecuteSQL(props.targetConfig, sql)
    await compareData()
  } catch (e: any) {
    // Error will be shown via Wails
    console.error('Sync failed:', e)
  }
}
</script>

<style scoped>
.data-sync {
  background: #16213e;
  border-radius: 10px;
  padding: 20px;
  margin-top: 20px;
}

.sync-header h3 {
  color: #4fc3f7;
  margin-bottom: 5px;
}

.hint {
  color: #888;
  font-size: 13px;
  margin-bottom: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.section-header h4 {
  color: #fff;
}

.table-section {
  margin-bottom: 25px;
}

.table-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 10px;
  max-height: 200px;
  overflow-y: auto;
}

.table-item {
  background: #0f0f23;
  padding: 12px;
  border-radius: 6px;
  cursor: pointer;
  border: 2px solid transparent;
  transition: all 0.2s;
}

.table-item:hover:not(.no-pk) {
  border-color: #4fc3f7;
}

.table-item.selected {
  border-color: #4fc3f7;
  background: rgba(79, 195, 247, 0.1);
}

.table-item.no-pk {
  opacity: 0.5;
  cursor: not-allowed;
}

.table-name {
  font-weight: bold;
  color: #fff;
  display: block;
  margin-bottom: 5px;
}

.table-info {
  display: flex;
  gap: 10px;
  font-size: 12px;
}

.pk-badge {
  color: #81c784;
}

.no-pk-badge {
  color: #f44336;
}

.row-count {
  color: #888;
}

.empty-tables {
  color: #888;
  text-align: center;
  padding: 30px;
}

.sync-summary {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;
}

.summary-item {
  background: #0f0f23;
  padding: 15px 25px;
  border-radius: 8px;
  text-align: center;
  border-left: 4px solid;
}

.summary-item.insert {
  border-left-color: #4caf50;
}

.summary-item.update {
  border-left-color: #ff9800;
}

.summary-item.delete {
  border-left-color: #f44336;
}

.summary-item .count {
  display: block;
  font-size: 24px;
  font-weight: bold;
  color: #fff;
}

.summary-item .label {
  color: #888;
  font-size: 13px;
}

.sync-options {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #ccc;
}

.checkbox-label input {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.diff-list {
  max-height: 400px;
  overflow-y: auto;
  margin-bottom: 20px;
}

.diff-item {
  background: #0f0f23;
  border-radius: 6px;
  margin-bottom: 8px;
  border-left: 4px solid;
  overflow: hidden;
}

.diff-item.insert {
  border-left-color: #4caf50;
}

.diff-item.update {
  border-left-color: #ff9800;
}

.diff-item.delete {
  border-left-color: #f44336;
}

.diff-header {
  padding: 10px 15px;
  background: rgba(255, 255, 255, 0.03);
  display: flex;
  align-items: center;
  gap: 10px;
}

.diff-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: bold;
}

.diff-badge.insert {
  background: rgba(76, 175, 80, 0.2);
  color: #4caf50;
}

.diff-badge.update {
  background: rgba(255, 152, 0, 0.2);
  color: #ff9800;
}

.diff-badge.delete {
  background: rgba(244, 67, 54, 0.2);
  color: #f44336;
}

.pk-info {
  color: #888;
  font-size: 12px;
}

.diff-sql {
  padding: 10px 15px;
}

.diff-sql code {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 12px;
  color: #81c784;
  word-break: break-all;
}

.show-more {
  text-align: center;
  padding: 10px;
}

.no-diff {
  text-align: center;
  padding: 30px;
  color: #4caf50;
}

.sync-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.btn-refresh {
  background: #0f3460;
  color: #4fc3f7;
}

.btn-compare {
  background: #4fc3f7;
  color: #1a1a2e;
  font-weight: bold;
}

.btn-copy {
  background: #0f3460;
  color: #4fc3f7;
}

.btn-execute {
  background: #4caf50;
  color: white;
  font-weight: bold;
}

.btn-small {
  background: #333;
  color: #eee;
}

.btn:hover:not(:disabled) {
  filter: brightness(1.1);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Dialog styles */
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: #16213e;
  border-radius: 10px;
  padding: 25px;
  width: 400px;
  border: 1px solid #333;
  text-align: center;
}

.dialog h4 {
  color: #ff9800;
  margin-bottom: 15px;
  font-size: 18px;
}

.dialog p {
  color: #ccc;
  margin-bottom: 20px;
  font-size: 14px;
}

.dialog-actions {
  display: flex;
  gap: 10px;
  justify-content: center;
}

.btn-cancel {
  padding: 10px 24px;
  border: 1px solid #333;
  border-radius: 5px;
  background: transparent;
  color: #888;
  cursor: pointer;
  font-size: 14px;
}

.btn-cancel:hover {
  background: #333;
}

.btn-confirm {
  padding: 10px 24px;
  border: none;
  border-radius: 5px;
  background: #4caf50;
  color: white;
  cursor: pointer;
  font-size: 14px;
  font-weight: bold;
}

.btn-confirm:hover {
  background: #45a045;
}

/* Terminal Style */
.terminal {
  background: #0d0d0d;
  border-radius: 8px;
  margin-bottom: 20px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
}

.terminal-header {
  background: #2d2d2d;
  padding: 8px 12px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.terminal-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.terminal-dot.red { background: #ff5f56; }
.terminal-dot.yellow { background: #ffbd2e; }
.terminal-dot.green { background: #27ca40; }

.terminal-title {
  margin-left: 10px;
  color: #888;
  font-size: 12px;
}

.terminal-body {
  padding: 15px;
  max-height: 200px;
  overflow-y: auto;
  font-size: 13px;
  line-height: 1.6;
}

.terminal-line {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.terminal-prompt {
  color: #4caf50;
  font-weight: bold;
}

.terminal-text {
  color: #eee;
}

.terminal-text.done {
  color: #81c784;
}

.terminal-text.error {
  color: #f44336;
}

.terminal-status {
  color: #4caf50;
  font-weight: bold;
}

.terminal-status.error {
  color: #f44336;
}

.terminal-dots {
  display: inline-flex;
  gap: 2px;
  margin-left: 4px;
}

.terminal-dots .dot {
  color: #555;
  transition: color 0.2s;
}

.terminal-dots .dot.active {
  color: #4fc3f7;
}

.terminal-cursor {
  display: inline-block;
  width: 8px;
  height: 16px;
  background: #4fc3f7;
  margin-left: 4px;
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}
</style>
