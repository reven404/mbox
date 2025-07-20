import FileProvider
import Foundation
import UniformTypeIdentifiers

/**
 * TusdFileProviderItem - Represents a file/folder item from tusd server
 * 
 * This class implements NSFileProviderItem protocol to provide file metadata
 * for items stored on the tusd server, enabling native Finder integration.
 */
class TusdFileProviderItem: NSObject, NSFileProviderItem {
    
    // MARK: - Core properties
    
    let itemIdentifier: NSFileProviderItemIdentifier
    let parentItemIdentifier: NSFileProviderItemIdentifier
    let filename: String
    let contentType: UTType
    
    // MARK: - File metadata
    
    private let _documentSize: NSNumber?
    private let _creationDate: Date?
    private let _contentModificationDate: Date?
    private let _isDownloaded: Bool
    private let _isShared: Bool
    private let _isUploaded: Bool
    private let _uploadingError: Error?
    private let _downloadingError: Error?
    
    // MARK: - Tusd-specific metadata
    
    private let tusdID: String?
    private let tusdOffset: Int64
    private let tusdMetadata: [String: String]
    
    // MARK: - Initialization
    
    init(identifier: NSFileProviderItemIdentifier,
         parentIdentifier: NSFileProviderItemIdentifier,
         filename: String,
         contentType: UTType,
         documentSize: NSNumber? = nil,
         creationDate: Date? = nil,
         contentModificationDate: Date? = nil,
         isDownloaded: Bool = false,
         isShared: Bool = false,
         isUploaded: Bool = true,
         uploadingError: Error? = nil,
         downloadingError: Error? = nil,
         tusdID: String? = nil,
         tusdOffset: Int64 = 0,
         tusdMetadata: [String: String] = [:]) {
        
        self.itemIdentifier = identifier
        self.parentItemIdentifier = parentIdentifier
        self.filename = filename
        self.contentType = contentType
        self._documentSize = documentSize
        self._creationDate = creationDate
        self._contentModificationDate = contentModificationDate
        self._isDownloaded = isDownloaded
        self._isShared = isShared
        self._isUploaded = isUploaded
        self._uploadingError = uploadingError
        self._downloadingError = downloadingError
        self.tusdID = tusdID
        self.tusdOffset = tusdOffset
        self.tusdMetadata = tusdMetadata
        
        super.init()
    }
    
    // MARK: - NSFileProviderItem protocol
    
    var documentSize: NSNumber? {
        return _documentSize
    }
    
    var creationDate: Date? {
        return _creationDate ?? Date()
    }
    
    var contentModificationDate: Date? {
        return _contentModificationDate ?? Date()
    }
    
    var lastUsedDate: Date? {
        return contentModificationDate
    }
    
    var tagData: Data? {
        return nil
    }
    
    var favoriteRank: NSNumber? {
        return nil
    }
    
    var isTrashed: Bool {
        return false
    }
    
    var isUploaded: Bool {
        return _isUploaded
    }
    
    var isUploading: Bool {
        return !_isUploaded && _uploadingError == nil
    }
    
    var uploadingError: Error? {
        return _uploadingError
    }
    
    var isDownloaded: Bool {
        return _isDownloaded
    }
    
    var isDownloading: Bool {
        return !_isDownloaded && _downloadingError == nil
    }
    
    var downloadingError: Error? {
        return _downloadingError
    }
    
    var isMostRecentVersionDownloaded: Bool {
        return _isDownloaded
    }
    
    var isShared: Bool {
        return _isShared
    }
    
    var isSharedByCurrentUser: Bool {
        return false
    }
    
    var ownerNameComponents: PersonNameComponents? {
        return nil
    }
    
    var mostRecentEditorNameComponents: PersonNameComponents? {
        return nil
    }
    
    var versionIdentifier: Data? {
        // Use tusd offset as version identifier
        return String(tusdOffset).data(using: .utf8)
    }
    
    var capabilities: NSFileProviderItemCapabilities {
        var caps: NSFileProviderItemCapabilities = []
        
        if contentType.conforms(to: .folder) {
            caps.insert([.allowsAddingSubItems, .allowsContentEnumerating])
        } else {
            caps.insert([.allowsReading, .allowsWriting])
        }
        
        caps.insert([.allowsRenaming, .allowsDeleting])
        
        return caps
    }
    
    // MARK: - Convenience methods
    
    var isFolder: Bool {
        return contentType.conforms(to: .folder)
    }
    
    var tusdUploadID: String? {
        return tusdID
    }
    
    var tusdUploadOffset: Int64 {
        return tusdOffset
    }
    
    var tusdFileMetadata: [String: String] {
        return tusdMetadata
    }
    
    // MARK: - Factory methods
    
    static func rootItem() -> TusdFileProviderItem {
        return TusdFileProviderItem(
            identifier: .rootContainer,
            parentIdentifier: .rootContainer,
            filename: "",
            contentType: .folder,
            isDownloaded: true,
            isUploaded: true
        )
    }
    
    static func folderItem(identifier: NSFileProviderItemIdentifier,
                          parentIdentifier: NSFileProviderItemIdentifier,
                          name: String) -> TusdFileProviderItem {
        return TusdFileProviderItem(
            identifier: identifier,
            parentIdentifier: parentIdentifier,
            filename: name,
            contentType: .folder,
            isDownloaded: true,
            isUploaded: true
        )
    }
    
    static func fileItem(identifier: NSFileProviderItemIdentifier,
                        parentIdentifier: NSFileProviderItemIdentifier,
                        name: String,
                        size: Int64,
                        modificationDate: Date,
                        tusdID: String? = nil,
                        tusdOffset: Int64 = 0) -> TusdFileProviderItem {
        
        let contentType = UTType(filenameExtension: (name as NSString).pathExtension) ?? .data
        
        return TusdFileProviderItem(
            identifier: identifier,
            parentIdentifier: parentIdentifier,
            filename: name,
            contentType: contentType,
            documentSize: NSNumber(value: size),
            contentModificationDate: modificationDate,
            isDownloaded: false, // Files need to be downloaded on demand
            isUploaded: true,
            tusdID: tusdID,
            tusdOffset: tusdOffset
        )
    }
}

// MARK: - NSFileProviderItemIdentifier extensions

extension NSFileProviderItemIdentifier {
    static func tusdIdentifier(for path: String) -> NSFileProviderItemIdentifier {
        // Create deterministic identifier from path
        let pathData = path.data(using: .utf8) ?? Data()
        let hash = pathData.base64EncodedString()
        return NSFileProviderItemIdentifier(hash)
    }
    
    var tusdPath: String {
        // Decode path from identifier
        guard let data = Data(base64Encoded: rawValue),
              let path = String(data: data, encoding: .utf8) else {
            return rawValue
        }
        return path
    }
}