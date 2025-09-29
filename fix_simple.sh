#!/bin/bash

# 简单修复HTTP处理器中的语法错误

# 查找所有HTTP处理器文件
find internal/interfaces/http -name "*.go" -type f | while read file; do
    echo "修复文件: $file"
    
    # 修复双逗号
    sed -i '' 's/,,/,/g' "$file"
    
    # 修复多余的括号
    sed -i '' 's/}}}}})/}})/g' "$file"
    sed -i '' 's/}}}}/}}/g' "$file"
    sed -i '' 's/}}}/})/g' "$file"
    
    # 修复Message字段为Error字段
    sed -i '' 's/Message: \&APIError{/Error: \&APIError{/g' "$file"
done

echo "修复完成"
