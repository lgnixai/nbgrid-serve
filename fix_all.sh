#!/bin/bash

# 全面修复HTTP处理器中的语法错误

# 查找所有HTTP处理器文件
find internal/interfaces/http -name "*.go" -type f | while read file; do
    echo "修复文件: $file"
    
    # 修复双逗号
    sed -i '' 's/,,/,/g' "$file"
    
    # 修复多余的括号
    sed -i '' 's/}}}}})/}})/g' "$file"
    sed -i '' 's/}}}}/}}/g' "$file"
    sed -i '' 's/}}}/})/g' "$file"
    sed -i '' 's/}}/})/g' "$file"
    
    # 修复Message字段为Error字段
    sed -i '' 's/Message: \&APIError{/Error: \&APIError{/g' "$file"
    
    # 修复TraceID字段位置
    sed -i '' 's/Code:    "[^"]*",$/Code:    "\1",/g' "$file"
    sed -i '' 's/Details: \([^}]*\),$/Details: \1,/g' "$file"
    
    # 修复结构体字面量
    sed -i '' 's/ErrorResponse{/APIResponse{Success: false, Error: \&APIError{/g' "$file"
    
    # 修复字段名
    sed -i '' 's/Error: /Message: /g' "$file"
    
    # 修复最后的闭合括号
    sed -i '' 's/})$/}})/g' "$file"
done

echo "修复完成"
