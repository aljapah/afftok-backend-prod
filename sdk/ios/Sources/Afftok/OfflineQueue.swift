import Foundation

/// Offline queue for storing events when device is offline
/// Supports persistence and automatic retry with exponential backoff
public class OfflineQueue {
    
    private var queue: [QueueItem] = []
    private let queueLock = NSLock()
    private var flushTimer: Timer?
    private var isProcessing = false
    private let debug: Bool
    
    public init(debug: Bool = false) {
        self.debug = debug
        loadFromStorage()
    }
    
    /// Add item to queue
    @discardableResult
    public func enqueue(type: String, payload: [String: Any]) -> String {
        queueLock.lock()
        defer { queueLock.unlock() }
        
        let id = UUID().uuidString
        
        // Convert payload to string dictionary for Codable
        var stringPayload: [String: String] = [:]
        for (key, value) in payload {
            if let stringValue = value as? String {
                stringPayload[key] = stringValue
            } else {
                stringPayload[key] = String(describing: value)
            }
        }
        
        let item = QueueItem(
            id: id,
            type: type,
            payload: stringPayload,
            timestamp: Date().timeIntervalSince1970,
            retryCount: 0,
            nextRetryTime: 0
        )
        
        if queue.count >= AfftokConfig.maxQueueSize {
            queue.removeFirst()
            log("Queue full, removed oldest item")
        }
        
        queue.append(item)
        saveToStorage()
        log("Enqueued \(type) event: \(id)")
        
        return id
    }
    
    /// Get all pending items ready for processing
    public func getPendingItems() -> [QueueItem] {
        queueLock.lock()
        defer { queueLock.unlock() }
        
        let now = Date().timeIntervalSince1970
        return queue.filter { $0.nextRetryTime <= now }
    }
    
    /// Mark item as completed and remove from queue
    public func markCompleted(id: String) {
        queueLock.lock()
        defer { queueLock.unlock() }
        
        queue.removeAll { $0.id == id }
        saveToStorage()
        log("Completed: \(id)")
    }
    
    /// Mark item for retry with exponential backoff
    public func markForRetry(id: String) {
        queueLock.lock()
        defer { queueLock.unlock() }
        
        guard let index = queue.firstIndex(where: { $0.id == id }) else { return }
        var item = queue[index]
        
        if item.retryCount >= AfftokConfig.maxRetryAttempts {
            queue.remove(at: index)
            log("Max retries reached, removing: \(id)")
            saveToStorage()
            return
        }
        
        // Calculate next retry time with exponential backoff + jitter
        let delay = min(
            AfftokConfig.initialRetryDelayMs * pow(2.0, Double(item.retryCount)),
            AfftokConfig.maxRetryDelayMs
        )
        let jitter = Double.random(in: 0...(delay * 0.1))
        
        item.retryCount += 1
        item.nextRetryTime = Date().timeIntervalSince1970 + delay + jitter
        
        queue[index] = item
        saveToStorage()
        
        log("Marked for retry (\(item.retryCount)/\(AfftokConfig.maxRetryAttempts)): \(id)")
    }
    
    /// Get queue size
    public func size() -> Int {
        queueLock.lock()
        defer { queueLock.unlock() }
        return queue.count
    }
    
    /// Check if queue is empty
    public func isEmpty() -> Bool {
        queueLock.lock()
        defer { queueLock.unlock() }
        return queue.isEmpty
    }
    
    /// Clear all items
    public func clear() {
        queueLock.lock()
        defer { queueLock.unlock() }
        
        queue.removeAll()
        saveToStorage()
        log("Queue cleared")
    }
    
    /// Start automatic flush timer
    public func startAutoFlush(interval: TimeInterval, processor: @escaping (QueueItem) async -> Bool) {
        stopAutoFlush()
        
        DispatchQueue.main.async { [weak self] in
            self?.flushTimer = Timer.scheduledTimer(withTimeInterval: interval, repeats: true) { [weak self] _ in
                Task {
                    await self?.flush(processor: processor)
                }
            }
        }
        log("Auto-flush started with interval: \(interval)s")
    }
    
    /// Stop automatic flush timer
    public func stopAutoFlush() {
        flushTimer?.invalidate()
        flushTimer = nil
        log("Auto-flush stopped")
    }
    
    /// Manually flush queue
    public func flush(processor: @escaping (QueueItem) async -> Bool) async {
        guard !isProcessing else {
            log("Flush already in progress, skipping")
            return
        }
        
        isProcessing = true
        log("Starting flush, \(queue.count) items in queue")
        
        let pendingItems = getPendingItems()
        
        for item in pendingItems {
            do {
                let success = await processor(item)
                if success {
                    markCompleted(id: item.id)
                } else {
                    markForRetry(id: item.id)
                }
            } catch {
                log("Error processing item \(item.id): \(error.localizedDescription)")
                markForRetry(id: item.id)
            }
        }
        
        isProcessing = false
        log("Flush completed, \(queue.count) items remaining")
    }
    
    /// Save queue to UserDefaults
    private func saveToStorage() {
        do {
            let data = try JSONEncoder().encode(queue)
            UserDefaults.standard.set(data, forKey: AfftokConfig.queueKey)
        } catch {
            log("Error saving queue: \(error.localizedDescription)")
        }
    }
    
    /// Load queue from UserDefaults
    private func loadFromStorage() {
        guard let data = UserDefaults.standard.data(forKey: AfftokConfig.queueKey) else { return }
        
        do {
            queue = try JSONDecoder().decode([QueueItem].self, from: data)
            log("Loaded \(queue.count) items from storage")
        } catch {
            log("Error loading queue: \(error.localizedDescription)")
        }
    }
    
    private func log(_ message: String) {
        if debug {
            print("[AffTok Queue] \(message)")
        }
    }
}

