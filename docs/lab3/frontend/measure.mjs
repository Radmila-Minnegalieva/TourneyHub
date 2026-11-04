import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import ts from '../../../frontend/node_modules/typescript/lib/typescript.js'

const totals = { files: 0, nonblankLines: 0, functions: 0, ternaries: 0, maxTernaryDepth: 0, logicalOperators: 0 }
function measure(directory) {
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) {
      measure(path)
      continue
    }
    if (!/\.tsx?$/.test(path) || path.endsWith('.d.ts')) continue
    const source = readFileSync(path, 'utf8')
    totals.files++
    totals.nonblankLines += source.split(/\r?\n/).filter(line => line.trim()).length
    const tree = ts.createSourceFile(path, source, ts.ScriptTarget.Latest)
    function visit(node, depth) {
      if (ts.isFunctionLike(node) && node.body) totals.functions++
      if (ts.isConditionalExpression(node)) {
        totals.ternaries++
        depth++
        totals.maxTernaryDepth = Math.max(totals.maxTernaryDepth, depth)
      }
      if (ts.isBinaryExpression(node) && [ts.SyntaxKind.AmpersandAmpersandToken, ts.SyntaxKind.BarBarToken].includes(node.operatorToken.kind)) totals.logicalOperators++
      ts.forEachChild(node, child => visit(child, depth))
    }
    visit(tree, 0)
  }
}
measure(process.argv[2] ?? 'frontend/src')
console.log(JSON.stringify(totals, null, 2))
