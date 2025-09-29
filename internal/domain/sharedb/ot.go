package sharedb

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// OTEngine 操作转换引擎
type OTEngine struct {
	types map[string]OTType
}

// OTType 操作转换类型
type OTType interface {
	// 应用操作到文档
	Apply(doc interface{}, op OTOperation) (interface{}, error)

	// 转换操作
	Transform(op1, op2 OTOperation) (OTOperation, OTOperation, error)

	// 检查操作是否有效
	Validate(op OTOperation, doc interface{}) error

	// 获取操作名称
	Name() string
}

// JSON0Type JSON0类型实现
type JSON0Type struct{}

// Name 返回类型名称
func (j *JSON0Type) Name() string {
	return "json0"
}

// Apply 应用操作到文档
func (j *JSON0Type) Apply(doc interface{}, op OTOperation) (interface{}, error) {
	if len(op.P) == 0 {
		return op.OI, nil
	}

	// 深拷贝文档
	docBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal document: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(docBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document: %w", err)
	}

	// 应用操作
	return j.applyOperation(result, op)
}

// applyOperation 递归应用操作
func (j *JSON0Type) applyOperation(doc interface{}, op OTOperation) (interface{}, error) {
	if len(op.P) == 0 {
		// 根级别操作
		if op.OI != nil && op.OD == nil {
			// 插入操作
			return op.OI, nil
		} else if op.OI == nil && op.OD != nil {
			// 删除操作
			return nil, nil
		} else if op.OI != nil && op.OD != nil {
			// 替换操作
			return op.OI, nil
		}
		return doc, nil
	}

	// 获取当前路径
	path := op.P[0]
	remainingPath := op.P[1:]

	// 根据文档类型处理
	switch v := doc.(type) {
	case map[string]interface{}:
		key, ok := path.(string)
		if !ok {
			return nil, fmt.Errorf("invalid path key type: %T", path)
		}

		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = val
		}

		if len(remainingPath) == 0 {
			// 到达目标路径
			if op.OI != nil && op.OD == nil {
				// 插入操作
				result[key] = op.OI
			} else if op.OI == nil && op.OD != nil {
				// 删除操作
				delete(result, key)
			} else if op.OI != nil && op.OD != nil {
				// 替换操作
				result[key] = op.OI
			}
		} else {
			// 继续递归
			if val, exists := result[key]; exists {
				newVal, err := j.applyOperation(val, OTOperation{
					P:  remainingPath,
					OI: op.OI,
					OD: op.OD,
				})
				if err != nil {
					return nil, err
				}
				result[key] = newVal
			} else if op.OI != nil {
				// 创建新路径
				newVal, err := j.applyOperation(nil, OTOperation{
					P:  remainingPath,
					OI: op.OI,
					OD: op.OD,
				})
				if err != nil {
					return nil, err
				}
				result[key] = newVal
			}
		}

		return result, nil

	case []interface{}:
		index, ok := path.(float64)
		if !ok {
			return nil, fmt.Errorf("invalid array index type: %T", path)
		}

		idx := int(index)
		if idx < 0 || idx >= len(v) {
			return nil, fmt.Errorf("array index out of bounds: %d", idx)
		}

		result := make([]interface{}, len(v))
		copy(result, v)

		if len(remainingPath) == 0 {
			// 到达目标路径
			if op.OI != nil && op.OD == nil {
				// 插入操作
				result = append(result[:idx], append([]interface{}{op.OI}, result[idx:]...)...)
			} else if op.OI == nil && op.OD != nil {
				// 删除操作
				result = append(result[:idx], result[idx+1:]...)
			} else if op.OI != nil && op.OD != nil {
				// 替换操作
				result[idx] = op.OI
			}
		} else {
			// 继续递归
			newVal, err := j.applyOperation(result[idx], OTOperation{
				P:  remainingPath,
				OI: op.OI,
				OD: op.OD,
			})
			if err != nil {
				return nil, err
			}
			result[idx] = newVal
		}

		return result, nil

	default:
		return nil, fmt.Errorf("unsupported document type: %T", doc)
	}
}

// Transform 转换操作
func (j *JSON0Type) Transform(op1, op2 OTOperation) (OTOperation, OTOperation, error) {
	// 检查路径是否相同
	if !j.pathsEqual(op1.P, op2.P) {
		// 路径不同，操作不冲突
		return op1, op2, nil
	}

	// 路径相同，需要转换
	if op1.OI != nil && op1.OD == nil && op2.OI != nil && op2.OD == nil {
		// 两个插入操作，第二个操作需要调整
		return op1, OTOperation{
			P:  j.incrementPath(op2.P),
			OI: op2.OI,
			OD: op2.OD,
		}, nil
	} else if op1.OI == nil && op1.OD != nil && op2.OI == nil && op2.OD != nil {
		// 两个删除操作，第二个操作需要调整
		return op1, OTOperation{
			P:  j.decrementPath(op2.P),
			OI: op2.OI,
			OD: op2.OD,
		}, nil
	} else if op1.OI != nil && op1.OD == nil && op2.OI == nil && op2.OD != nil {
		// 插入和删除操作，删除操作需要调整
		return op1, OTOperation{
			P:  j.decrementPath(op2.P),
			OI: op2.OI,
			OD: op2.OD,
		}, nil
	} else if op1.OI == nil && op1.OD != nil && op2.OI != nil && op2.OD == nil {
		// 删除和插入操作，插入操作需要调整
		return op1, OTOperation{
			P:  j.incrementPath(op2.P),
			OI: op2.OI,
			OD: op2.OD,
		}, nil
	}

	// 其他情况，操作不冲突
	return op1, op2, nil
}

// Validate 验证操作
func (j *JSON0Type) Validate(op OTOperation, doc interface{}) error {
	// 检查路径有效性
	if len(op.P) == 0 {
		return nil
	}

	// 检查操作类型
	if op.OI == nil && op.OD == nil {
		return fmt.Errorf("operation must have either OI or OD")
	}

	// 检查路径是否可以访问
	return j.validatePath(doc, op.P)
}

// pathsEqual 检查路径是否相等
func (j *JSON0Type) pathsEqual(p1, p2 []interface{}) bool {
	if len(p1) != len(p2) {
		return false
	}

	for i, v1 := range p1 {
		if !reflect.DeepEqual(v1, p2[i]) {
			return false
		}
	}

	return true
}

// incrementPath 增加路径
func (j *JSON0Type) incrementPath(path []interface{}) []interface{} {
	if len(path) == 0 {
		return path
	}

	result := make([]interface{}, len(path))
	copy(result, path)

	// 如果是数字索引，增加1
	if index, ok := result[len(result)-1].(float64); ok {
		result[len(result)-1] = index + 1
	}

	return result
}

// decrementPath 减少路径
func (j *JSON0Type) decrementPath(path []interface{}) []interface{} {
	if len(path) == 0 {
		return path
	}

	result := make([]interface{}, len(path))
	copy(result, path)

	// 如果是数字索引，减少1
	if index, ok := result[len(result)-1].(float64); ok && index > 0 {
		result[len(result)-1] = index - 1
	}

	return result
}

// validatePath 验证路径
func (j *JSON0Type) validatePath(doc interface{}, path []interface{}) error {
	if len(path) == 0 {
		return nil
	}

	current := doc
	for i, p := range path {
		switch v := current.(type) {
		case map[string]interface{}:
			key, ok := p.(string)
			if !ok {
				return fmt.Errorf("invalid path key type at index %d: %T", i, p)
			}
			current = v[key]
		case []interface{}:
			index, ok := p.(float64)
			if !ok {
				return fmt.Errorf("invalid array index type at index %d: %T", i, p)
			}
			idx := int(index)
			if idx < 0 || idx >= len(v) {
				return fmt.Errorf("array index out of bounds at index %d: %d", i, idx)
			}
			current = v[idx]
		default:
			return fmt.Errorf("cannot access path at index %d: unsupported type %T", i, current)
		}
	}

	return nil
}

// NewOTEngine 创建操作转换引擎
func NewOTEngine() *OTEngine {
	engine := &OTEngine{
		types: make(map[string]OTType),
	}

	// 注册默认类型
	engine.RegisterType(&JSON0Type{})

	return engine
}

// RegisterType 注册操作转换类型
func (e *OTEngine) RegisterType(otType OTType) {
	e.types[otType.Name()] = otType
}

// GetType 获取操作转换类型
func (e *OTEngine) GetType(name string) (OTType, error) {
	otType, exists := e.types[name]
	if !exists {
		return nil, fmt.Errorf("unknown OT type: %s", name)
	}
	return otType, nil
}

// ApplyOperation 应用操作
func (e *OTEngine) ApplyOperation(doc interface{}, op OTOperation, typeName string) (interface{}, error) {
	otType, err := e.GetType(typeName)
	if err != nil {
		return nil, err
	}

	return otType.Apply(doc, op)
}

// TransformOperations 转换操作
func (e *OTEngine) TransformOperations(op1, op2 OTOperation, typeName string) (OTOperation, OTOperation, error) {
	otType, err := e.GetType(typeName)
	if err != nil {
		return OTOperation{}, OTOperation{}, err
	}

	return otType.Transform(op1, op2)
}

// ValidateOperation 验证操作
func (e *OTEngine) ValidateOperation(op OTOperation, doc interface{}, typeName string) error {
	otType, err := e.GetType(typeName)
	if err != nil {
		return err
	}

	return otType.Validate(op, doc)
}
