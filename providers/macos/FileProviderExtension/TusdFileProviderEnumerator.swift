import FileProvider
import Foundation
import os.log

/**
 * TusdFileProviderEnumerator - Enumerates files and folders from tusd server
 * 
 * This class handles listing directory contents from the tusd server
 * and provides them to the File Provider system for display in Finder.
 */
class TusdFileProviderEnumerator: NSObject, NSFileProviderEnumerator {
    
    private let enumeratedItemIdentifier: NSFileProviderItemIdentifier
    private let client: TusdFileProviderClient
    private let logger: Logger
    private var anchor: NSFileProviderSyncAnchor?
    
    init(enumeratedItemIdentifier: NSFileProviderItemIdentifier,
         client: TusdFileProviderClient,
         logger: Logger) {
        self.enumeratedItemIdentifier = enumeratedItemIdentifier
        self.client = client
        self.logger = logger
        super.init()
    }
    
    // MARK: - NSFileProviderEnumerator
    
    func invalidate() {
        logger.info("Enumerator invalidated for: \(enumeratedItemIdentifier.rawValue)")
    }
    
    func enumerateItems(for observer: NSFileProviderEnumerationObserver, startingAt page: NSFileProviderPage) {
        logger.info("Enumerating items for: \(enumeratedItemIdentifier.rawValue), page: \(page.rawValue)")
        
        Task {
            do {
                let items = try await client.listItems(in: enumeratedItemIdentifier, startingAt: page)
                
                // Create sync anchor for tracking changes
                let newAnchor = NSFileProviderSyncAnchor(Date().timeIntervalSince1970.description.data(using: .utf8)!)
                
                if items.isEmpty {
                    observer.finishEnumerating(upTo: nil)
                } else {
                    observer.didEnumerate(items)
                    observer.finishEnumerating(upTo: nil)
                }
                
                self.anchor = newAnchor
                
            } catch {
                logger.error("Enumeration failed: \(error.localizedDescription)")
                observer.finishEnumeratingWithError(error)
            }
        }
    }
    
    func enumerateChanges(for observer: NSFileProviderChangeObserver, from anchor: NSFileProviderSyncAnchor) {
        logger.info("Enumerating changes from anchor for: \(enumeratedItemIdentifier.rawValue)")
        
        Task {
            do {
                let (updatedItems, deletedItems, newAnchor) = try await client.getChanges(
                    in: enumeratedItemIdentifier,
                    since: anchor
                )
                
                if !updatedItems.isEmpty {
                    observer.didUpdate(updatedItems)
                }
                
                if !deletedItems.isEmpty {
                    observer.didDeleteItems(withIdentifiers: deletedItems)
                }
                
                observer.finishEnumeratingChanges(upTo: newAnchor, moreComing: false)
                self.anchor = newAnchor
                
            } catch {
                logger.error("Change enumeration failed: \(error.localizedDescription)")
                observer.finishEnumeratingWithError(error)
            }
        }
    }
    
    func currentSyncAnchor(completionHandler: @escaping (NSFileProviderSyncAnchor?) -> Void) {
        completionHandler(anchor)
    }
}

/**
 * TusdWorkingSetEnumerator - Manages the working set of actively used files
 * 
 * The working set contains files that are currently being accessed or modified,
 * allowing the system to prioritize their synchronization and caching.
 */
class TusdWorkingSetEnumerator: NSObject, NSFileProviderEnumerator {
    
    private let workingSet: NSFileProviderWorkingSet
    private let logger: Logger
    
    init(workingSet: NSFileProviderWorkingSet, logger: Logger) {
        self.workingSet = workingSet
        self.logger = logger
        super.init()
    }
    
    func invalidate() {
        logger.info("Working set enumerator invalidated")
    }
    
    func enumerateItems(for observer: NSFileProviderEnumerationObserver, startingAt page: NSFileProviderPage) {
        logger.info("Enumerating working set items")
        
        // For now, return empty working set
        // In a full implementation, this would track actively used files
        observer.finishEnumeration(upTo: nil)
    }
    
    func enumerateChanges(for observer: NSFileProviderChangeObserver, from anchor: NSFileProviderSyncAnchor) {
        logger.info("Enumerating working set changes")
        
        // Create new anchor
        let newAnchor = NSFileProviderSyncAnchor(Date().timeIntervalSince1970.description.data(using: .utf8)!)
        observer.finishEnumeratingChanges(upTo: newAnchor, moreComing: false)
    }
    
    func currentSyncAnchor(completionHandler: @escaping (NSFileProviderSyncAnchor?) -> Void) {
        let anchor = NSFileProviderSyncAnchor(Date().timeIntervalSince1970.description.data(using: .utf8)!)
        completionHandler(anchor)
    }
}

// MARK: - Helper extensions

extension NSFileProviderPage {
    static var initial: NSFileProviderPage {
        return NSFileProviderPage(Data())
    }
}

extension NSFileProviderSyncAnchor {
    var timestamp: TimeInterval? {
        guard let data = self as? Data,
              let string = String(data: data, encoding: .utf8),
              let timestamp = TimeInterval(string) else {
            return nil
        }
        return timestamp
    }
    
    static func current() -> NSFileProviderSyncAnchor {
        let timestamp = Date().timeIntervalSince1970.description
        return NSFileProviderSyncAnchor(timestamp.data(using: .utf8)!)
    }
}