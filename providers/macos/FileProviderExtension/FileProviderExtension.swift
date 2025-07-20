import FileProvider
import Foundation
import os.log

/**
 * TusdFileProviderExtension - Real File Provider implementation for tusd integration
 * 
 * This extension provides native macOS Finder integration for tusd file servers,
 * allowing users to access remote files as if they were local through the File Provider API.
 */
class TusdFileProviderExtension: NSFileProviderExtension {
    
    private let logger = Logger(subsystem: "com.mdriver.fileprovider", category: "FileProvider")
    private var tusdClient: TusdFileProviderClient?
    private let workingSet = NSFileProviderWorkingSet()
    
    override init() {
        super.init()
        logger.info("TusdFileProviderExtension initialized")
        setupTusdClient()
    }
    
    private func setupTusdClient() {
        // Initialize connection to Go backend via CGO
        if let serverURL = UserDefaults.standard.string(forKey: "TusdServerURL") {
            tusdClient = TusdFileProviderClient(serverURL: serverURL)
            logger.info("TusdClient configured for server: \(serverURL)")
        } else {
            logger.error("TusdServerURL not configured in UserDefaults")
        }
    }
    
    // MARK: - Enumeration
    
    override func enumerator(for containerItemIdentifier: NSFileProviderItemIdentifier) throws -> NSFileProviderEnumerator {
        logger.info("Creating enumerator for container: \(containerItemIdentifier.rawValue)")
        
        guard let client = tusdClient else {
            throw NSFileProviderError(.notAuthenticated)
        }
        
        return TusdFileProviderEnumerator(
            enumeratedItemIdentifier: containerItemIdentifier,
            client: client,
            logger: logger
        )
    }
    
    // MARK: - Item retrieval
    
    override func item(for identifier: NSFileProviderItemIdentifier) throws -> NSFileProviderItem {
        logger.info("Fetching item for identifier: \(identifier.rawValue)")
        
        guard let client = tusdClient else {
            throw NSFileProviderError(.notAuthenticated)
        }
        
        return try client.getItem(for: identifier)
    }
    
    // MARK: - File content
    
    override func fetchContents(for itemIdentifier: NSFileProviderItemIdentifier, 
                               version requestedVersion: NSFileProviderItemVersion?, 
                               completionHandler: @escaping (URL?, NSFileProviderItem?, Error?) -> Void) {
        logger.info("Fetching contents for item: \(itemIdentifier.rawValue)")
        
        guard let client = tusdClient else {
            completionHandler(nil, nil, NSFileProviderError(.notAuthenticated))
            return
        }
        
        Task {
            do {
                let (tempURL, item) = try await client.fetchContents(for: itemIdentifier)
                completionHandler(tempURL, item, nil)
            } catch {
                self.logger.error("Failed to fetch contents: \(error.localizedDescription)")
                completionHandler(nil, nil, error)
            }
        }
    }
    
    // MARK: - File uploads
    
    override func createItem(basedOn itemTemplate: NSFileProviderItem, 
                           fields: NSFileProviderItemFields, 
                           contents url: URL?, 
                           options: NSFileProviderCreateItemOptions = [], 
                           completionHandler: @escaping (NSFileProviderItem?, NSFileProviderItemFields, Bool, Error?) -> Void) {
        
        logger.info("Creating item: \(itemTemplate.filename)")
        
        guard let client = tusdClient else {
            completionHandler(nil, [], false, NSFileProviderError(.notAuthenticated))
            return
        }
        
        guard let contentsURL = url else {
            completionHandler(nil, [], false, NSFileProviderError(.noSuchItem))
            return
        }
        
        Task {
            do {
                let newItem = try await client.createItem(
                    basedOn: itemTemplate,
                    contents: contentsURL
                )
                completionHandler(newItem, [], true, nil)
            } catch {
                self.logger.error("Failed to create item: \(error.localizedDescription)")
                completionHandler(nil, [], false, error)
            }
        }
    }
    
    // MARK: - Item modification
    
    override func modifyItem(_ item: NSFileProviderItem, 
                           baseVersion version: NSFileProviderItemVersion, 
                           changedFields: NSFileProviderItemFields, 
                           contents newContents: URL?, 
                           options: NSFileProviderModifyItemOptions = [], 
                           completionHandler: @escaping (NSFileProviderItem?, NSFileProviderItemFields, Bool, Error?) -> Void) {
        
        logger.info("Modifying item: \(item.filename)")
        
        guard let client = tusdClient else {
            completionHandler(nil, [], false, NSFileProviderError(.notAuthenticated))
            return
        }
        
        Task {
            do {
                let modifiedItem = try await client.modifyItem(
                    item,
                    changedFields: changedFields,
                    contents: newContents
                )
                completionHandler(modifiedItem, [], true, nil)
            } catch {
                self.logger.error("Failed to modify item: \(error.localizedDescription)")
                completionHandler(nil, [], false, error)
            }
        }
    }
    
    // MARK: - Item deletion
    
    override func deleteItem(withIdentifier itemIdentifier: NSFileProviderItemIdentifier, 
                           completionHandler: @escaping (Error?) -> Void) {
        logger.info("Deleting item: \(itemIdentifier.rawValue)")
        
        guard let client = tusdClient else {
            completionHandler(NSFileProviderError(.notAuthenticated))
            return
        }
        
        Task {
            do {
                try await client.deleteItem(withIdentifier: itemIdentifier)
                completionHandler(nil)
            } catch {
                self.logger.error("Failed to delete item: \(error.localizedDescription)")
                completionHandler(error)
            }
        }
    }
    
    // MARK: - Working set management
    
    override func enumeratorForWorkingSet() throws -> NSFileProviderEnumerator {
        logger.info("Creating working set enumerator")
        return TusdWorkingSetEnumerator(workingSet: workingSet, logger: logger)
    }
    
    // MARK: - Domain state
    
    override func supportedServiceSources(for itemIdentifier: NSFileProviderItemIdentifier, 
                                        completionHandler: @escaping ([NSFileProviderServiceSource]?, Error?) -> Void) {
        // Return custom service sources if needed
        completionHandler([], nil)
    }
}

// MARK: - Error handling extension

extension NSFileProviderError {
    static func tusdError(_ message: String) -> NSFileProviderError {
        return NSFileProviderError(.serverUnreachable, userInfo: [
            NSLocalizedDescriptionKey: message
        ])
    }
}