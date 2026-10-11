use tantivy::{
    Index,
    directory::MmapDirectory,
    schema::{STORED, Schema, TEXT},
};

fn main() {
    let mut schema_builder = Schema::builder();

    let title = schema_builder.add_text_field("title", TEXT | STORED);
    let description = schema_builder.add_text_field("description", TEXT | STORED);
    let body = schema_builder.add_text_field("body", TEXT);

    let schema = schema_builder.build();
    let index_directory = MmapDirectory::open("/var/index")?;
    let index = Index::open_or_create(index_directory, schema);
}
