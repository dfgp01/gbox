package ecs

// 系统调度（SystemGroup）已上移到 pkg/game：维护一组 System、按优先级降序稳定排序并执行。
// internal 不再保留调度器默认实现；业务系统由 pkg/game.SystemGroup 统一调度。
