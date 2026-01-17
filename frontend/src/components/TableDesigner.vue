<template>
  <div class="table-designer">
    <div class="designer-header">
      <h3>🛠️ Table Designer</h3>
      <p class="hint">Design and create new tables</p>
    </div>

    <!-- Table Basic Info -->
    <div class="section">
      <h4>Table Info</h4>
      <div class="form-row">
        <div class="form-group">
          <label>Table Name</label>
          <input type="text" v-model="tableDef.name" placeholder="table_name" />
        </div>
        <div class="form-group">
          <label>Engine</label>
          <select v-model="tableDef.engine">
            <option v-for="e in engines" :key="e" :value="e">{{ e }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>Charset</label>
          <select v-model="tableDef.charset">
            <option v-for="c in charsets" :key="c" :value="c">{{ c }}</option>
          </select>
        </div>
      </div>
      <div class="form-row">
        <div class="form-group full-width">
          <label>Comment</label>
          <input type="text" v-model="tableDef.comment" placeholder="Table description (optional)" />
        </div>
      </div>
    </div>

    <!-- Columns -->
    <div class="section">
      <div class="section-header">
        <h4>Columns</h4>
        <button class="btn btn-add" @click="addColumn">+ Add Column</button>
      </div>

      <div class="columns-table">
        <div class="table-header">
          <span class="col-name">Name</span>
          <span class="col-type">Type</span>
          <span class="col-length">Length</span>
          <span class="col-nullable">Nullable</span>
          <span class="col-default">Default</span>
          <span class="col-pk">PK</span>
          <span class="col-ai">AI</span>
          <span class="col-actions">Actions</span>
        </div>

        <div
          v-for="(col, index) in tableDef.columns"
          :key="index"
          class="table-row"
        >
          <input class="col-name" v-model="col.name" placeholder="column_name" />
          <select class="col-type" v-model="col.type">
            <option v-for="t in dataTypes" :key="t" :value="t">{{ t }}</option>
          </select>
          <input class="col-length" type="number" v-model.number="col.length" placeholder="255" />
          <input class="col-nullable" type="checkbox" v-model="col.nullable" />
          <input class="col-default" v-model="col.defaultValue" placeholder="NULL" />
          <input class="col-pk" type="checkbox" v-model="col.primaryKey" @change="onPKChange(col)" />
          <input class="col-ai" type="checkbox" v-model="col.autoIncrement" :disabled="!col.primaryKey" />
          <div class="col-actions">
            <button class="btn-icon" @click="moveColumn(index, -1)" :disabled="index === 0">↑</button>
            <button class="btn-icon" @click="moveColumn(index, 1)" :disabled="index === tableDef.columns.length - 1">↓</button>
            <button class="btn-icon btn-delete" @click="removeColumn(index)">×</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Indexes -->
    <div class="section">
      <div class="section-header">
        <h4>Indexes</h4>
        <button class="btn btn-add" @click="addIndex">+ Add Index</button>
      </div>

      <div class="indexes-list" v-if="tableDef.indexes.length > 0">
        <div v-for="(idx, index) in tableDef.indexes" :key="index" class="index-item">
          <input v-model="idx.name" placeholder="index_name" class="index-name" />
          <select v-model="idx.columns" multiple class="index-columns">
            <option v-for="col in tableDef.columns" :key="col.name" :value="col.name">
              {{ col.name }}
            </option>
          </select>
          <label class="checkbox-label">
            <input type="checkbox" v-model="idx.unique" />
            <span>Unique</span>
          </label>
          <button class="btn-icon btn-delete" @click="removeIndex(index)">×</button>
        </div>
      </div>
      <div v-else class="no-indexes">No indexes defined</div>
    </div>

    <!-- SQL Preview -->
    <div class="section">
      <div class="section-header">
        <h4>SQL Preview</h4>
        <button class="btn btn-copy" @click="copySQL">📋 Copy</button>
      </div>
      <pre class="sql-preview"><code>{{ generatedSQL }}</code></pre>
    </div>

    <!-- Actions -->
    <div class="designer-actions">
      <button class="btn btn-reset" @click="resetForm">🔄 Reset</button>
      <button
        class="btn btn-execute"
        @click="createTable"
        :disabled="!canCreate || creating"
      >
        {{ creating ? 'Creating...' : '▶️ Create Table' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { GetCommonDataTypes, GetTableEngines, GetCharsets, GenerateCreateTableSQL, CreateTable } from '../../wailsjs/go/main/App'
import { database } from '../../wailsjs/go/models'

type ConnectionConfig = database.ConnectionConfig

// Define local interfaces for form state (avoid Wails class requirements)
interface ColumnDef {
  name: string
  type: string
  length: number
  nullable: boolean
  defaultValue: string | undefined
  autoIncrement: boolean
  primaryKey: boolean
  unique: boolean
  comment: string
}

interface IndexDef {
  name: string
  columns: string[]
  unique: boolean
}

interface TableDef {
  name: string
  columns: ColumnDef[]
  indexes: IndexDef[]
  engine: string
  charset: string
  collation: string
  comment: string
}

const props = defineProps<{
  targetConfig: ConnectionConfig
  targetConnected: boolean
}>()

// Data
const dataTypes = ref<string[]>([])
const engines = ref<string[]>([])
const charsets = ref<string[]>([])
const creating = ref(false)
const generatedSQL = ref('')

const tableDef = ref<TableDef>({
  name: '',
  columns: [],
  indexes: [],
  engine: 'InnoDB',
  charset: 'utf8mb4',
  collation: 'utf8mb4_unicode_ci',
  comment: ''
})

// Computed
const canCreate = computed(() => {
  return props.targetConnected &&
         props.targetConfig.database &&
         tableDef.value.name &&
         tableDef.value.columns.length > 0 &&
         tableDef.value.columns.every(c => c.name && c.type)
})

// Methods
onMounted(async () => {
  try {
    dataTypes.value = await GetCommonDataTypes()
    engines.value = await GetTableEngines()
    charsets.value = await GetCharsets()
  } catch (e) {
    console.error('Failed to load options:', e)
    // Fallback values
    dataTypes.value = ['INT', 'BIGINT', 'VARCHAR', 'TEXT', 'DATETIME', 'BOOLEAN']
    engines.value = ['InnoDB', 'MyISAM']
    charsets.value = ['utf8mb4', 'utf8']
  }

  // Add initial column
  addColumn()
})

// Watch for changes and regenerate SQL
watch(tableDef, async () => {
  if (tableDef.value.name && tableDef.value.columns.length > 0) {
    try {
      generatedSQL.value = await GenerateCreateTableSQL(tableDef.value as any)
    } catch (e) {
      generatedSQL.value = '-- Error generating SQL'
    }
  } else {
    generatedSQL.value = '-- Define table name and columns to preview SQL'
  }
}, { deep: true })

function addColumn() {
  tableDef.value.columns.push({
    name: '',
    type: 'VARCHAR',
    length: 255,
    nullable: true,
    defaultValue: undefined,
    autoIncrement: false,
    primaryKey: false,
    unique: false,
    comment: ''
  })
}

function removeColumn(index: number) {
  tableDef.value.columns.splice(index, 1)
}

function moveColumn(index: number, direction: number) {
  const newIndex = index + direction
  if (newIndex < 0 || newIndex >= tableDef.value.columns.length) return

  const cols = tableDef.value.columns
  const temp = cols[index]
  cols[index] = cols[newIndex]
  cols[newIndex] = temp
}

function onPKChange(col: ColumnDef) {
  if (col.primaryKey) {
    col.nullable = false
  } else {
    col.autoIncrement = false
  }
}

function addIndex() {
  tableDef.value.indexes.push({
    name: `idx_${tableDef.value.name || 'table'}_${tableDef.value.indexes.length + 1}`,
    columns: [],
    unique: false
  })
}

function removeIndex(index: number) {
  tableDef.value.indexes.splice(index, 1)
}

function copySQL() {
  navigator.clipboard.writeText(generatedSQL.value)
  alert('SQL copied to clipboard!')
}

function resetForm() {
  tableDef.value = {
    name: '',
    columns: [],
    indexes: [],
    engine: 'InnoDB',
    charset: 'utf8mb4',
    collation: 'utf8mb4_unicode_ci',
    comment: ''
  }
  addColumn()
}

async function createTable() {
  if (!canCreate.value) return

  if (!confirm(`Create table "${tableDef.value.name}" in database "${props.targetConfig.database}"?`)) {
    return
  }

  creating.value = true
  try {
    await CreateTable(props.targetConfig, tableDef.value as any)
    alert(`Table "${tableDef.value.name}" created successfully!`)
    resetForm()
  } catch (e: any) {
    alert('Failed to create table: ' + e)
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.table-designer {
  background: #16213e;
  border-radius: 10px;
  padding: 20px;
  margin-top: 20px;
}

.designer-header h3 {
  color: #4fc3f7;
  margin-bottom: 5px;
}

.hint {
  color: #888;
  font-size: 13px;
  margin-bottom: 20px;
}

.section {
  margin-bottom: 25px;
}

.section h4 {
  color: #fff;
  margin-bottom: 15px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.form-row {
  display: flex;
  gap: 15px;
  margin-bottom: 10px;
}

.form-group {
  flex: 1;
}

.form-group.full-width {
  flex: 3;
}

.form-group label {
  display: block;
  font-size: 12px;
  color: #888;
  margin-bottom: 5px;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid #333;
  border-radius: 5px;
  background: #0f0f23;
  color: #eee;
  font-size: 14px;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #4fc3f7;
}

/* Columns Table */
.columns-table {
  background: #0f0f23;
  border-radius: 8px;
  overflow: hidden;
}

.table-header,
.table-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 15px;
}

.table-header {
  background: rgba(255, 255, 255, 0.05);
  font-size: 12px;
  color: #888;
  font-weight: bold;
}

.table-row {
  border-top: 1px solid #333;
}

.table-row input,
.table-row select {
  padding: 6px 8px;
  border: 1px solid #333;
  border-radius: 4px;
  background: #1a1a2e;
  color: #eee;
  font-size: 13px;
}

.table-row input:focus,
.table-row select:focus {
  outline: none;
  border-color: #4fc3f7;
}

.table-row input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.col-name { flex: 2; }
.col-type { flex: 1.5; }
.col-length { flex: 0.8; }
.col-nullable { flex: 0.5; justify-content: center; }
.col-default { flex: 1.2; }
.col-pk { flex: 0.4; justify-content: center; }
.col-ai { flex: 0.4; justify-content: center; }
.col-actions { flex: 1; display: flex; gap: 5px; justify-content: flex-end; }

/* Indexes */
.indexes-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.index-item {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #0f0f23;
  padding: 10px 15px;
  border-radius: 6px;
}

.index-name {
  flex: 1;
  padding: 6px 8px;
  border: 1px solid #333;
  border-radius: 4px;
  background: #1a1a2e;
  color: #eee;
}

.index-columns {
  flex: 2;
  padding: 6px 8px;
  border: 1px solid #333;
  border-radius: 4px;
  background: #1a1a2e;
  color: #eee;
  min-height: 60px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 5px;
  color: #ccc;
  font-size: 13px;
}

.no-indexes {
  color: #666;
  font-style: italic;
  padding: 15px;
  text-align: center;
}

/* SQL Preview */
.sql-preview {
  background: #0f0f23;
  padding: 15px;
  border-radius: 8px;
  overflow-x: auto;
  max-height: 200px;
}

.sql-preview code {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 13px;
  color: #81c784;
  white-space: pre-wrap;
}

/* Buttons */
.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.btn-add {
  background: #0f3460;
  color: #4fc3f7;
}

.btn-copy {
  background: #0f3460;
  color: #4fc3f7;
}

.btn-reset {
  background: #333;
  color: #ccc;
}

.btn-execute {
  background: #4caf50;
  color: white;
  font-weight: bold;
}

.btn:hover:not(:disabled) {
  filter: brightness(1.1);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-icon {
  padding: 4px 8px;
  border: none;
  border-radius: 4px;
  background: #333;
  color: #ccc;
  cursor: pointer;
  font-size: 14px;
}

.btn-icon:hover:not(:disabled) {
  background: #444;
}

.btn-icon:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.btn-delete {
  color: #f44336;
}

.btn-delete:hover {
  background: rgba(244, 67, 54, 0.2);
}

.designer-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #333;
}
</style>
