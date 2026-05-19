# TODO

- Move validation out of use-cases into the domain layer.
- Add a `ChatUpdated` event and a pipeline for delivering it to users.
- Add Merge() method to ValidationError. The method must add violations of the passed error.